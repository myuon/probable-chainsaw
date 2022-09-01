package service

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os/exec"
	"strings"
	"time"
)

type AnalyzeCommitService struct {
	commitRepository               infra.CommitRepository
	deployCommitRepository         infra.DeployCommitRepository
	deployCommitRelationRepository infra.DeployCommitRelationRepository
}

func NewAnalyzeService(project model.Project) (AnalyzeCommitService, error) {
	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return AnalyzeCommitService{}, err
	}

	commitRepository := infra.CommitRepository{Db: db}
	deployCommitRepository := infra.DeployCommitRepository{Db: db}
	deployCommitRelationRepository := infra.DeployCommitRelationRepository{Db: db}

	return AnalyzeCommitService{
		commitRepository:               commitRepository,
		deployCommitRepository:         deployCommitRepository,
		deployCommitRelationRepository: deployCommitRelationRepository,
	}, nil
}

func (service AnalyzeCommitService) ResetTables() error {
	if err := service.commitRepository.ResetTable(); err != nil {
		return err
	}
	if err := service.deployCommitRepository.ResetTable(); err != nil {
		return err
	}
	if err := service.deployCommitRelationRepository.ResetTable(); err != nil {
		return err
	}

	return nil
}

func (service AnalyzeCommitService) UpdateRepositoryCommits(p model.ProjectRepository) error {
	repo, err := infra.GitOperatorCloneOrPull(p.WorkPath(), fmt.Sprintf("git@github.com:%v/%v.git", p.Org, p.Name))
	if err != nil {
		return err
	}

	// save commits from HEAD
	commits, err := repo.GetCommitsFromHEAD()
	if err != nil {
		return err
	}

	if err := service.commitRepository.Save(p.RepositoryName(), commits); err != nil {
		return err
	}

	// find deployed commits from "deploy" branch
	commits, err = repo.GetCommitsInBranch("origin/deploy")
	if err != nil {
		return err
	}

	deployCommits := []model.DeployCommit{}
	if err := commits.ForEach(func(c *object.Commit) error {
		// check if this is a merge commit from "master" branch
		// FIXME: filter only `Merge pull request #XXX from NAME/BRANCH` ones
		if !(strings.Contains(c.Message, "Merge pull request") && strings.Contains(c.Message, "master")) {
			return nil
		}

		previous := ""
		if c.NumParents() > 0 {
			parent, err := c.Parent(0)
			if err != nil {
				return err
			}
			previous = parent.Hash.String()
		}

		deployCommits = append(deployCommits, model.DeployCommit{
			Hash:           c.Hash.String(),
			AuthorName:     c.Author.Name,
			DeployedAt:     c.Author.When.Unix(),
			PreviousHash:   previous,
			RepositoryName: p.RepositoryName(),
		})

		return nil
	}); err != nil {
		return err
	}

	if err := service.deployCommitRepository.Create(deployCommits); err != nil {
		return err
	}

	if err := service.commitRepository.UpdateDeployTags("master", commits); err != nil {
		return err
	}

	return nil
}

func (service AnalyzeCommitService) UpdateDeployCommitRelationsOver(p model.ProjectRepository, startDate time.Time, endDate time.Time) error {
	current := startDate

	for current.Before(endDate) {
		deploys, err := service.deployCommitRepository.FindBetweenDeployedAt(current.Unix(), current.Add(24*time.Hour).Unix())
		if err != nil {
			return err
		}

		for _, d := range deploys {
			bin, err := exec.Command("git", "-C", p.WorkPath(), "log", "--pretty=format:%H", fmt.Sprintf("%v..%v", d.PreviousHash, d.Hash)).Output()
			if err != nil {
				return err
			}

			relations := []infra.DeployCommitRelation{}
			for _, hash := range strings.Split(string(bin), "\n") {
				if hash == "" {
					continue
				}
				if d.Hash == hash {
					continue
				}

				relations = append(relations, infra.DeployCommitRelation{
					DeployHash: d.Hash,
					CommitHash: hash,
				})
			}

			if err := service.deployCommitRelationRepository.Save(relations); err != nil {
				return err
			}

			log.Info().Msgf("%v", string(bin))
		}

		log.Info().Int(fmt.Sprintf("%v (%v)", current.Format("2006-01-02"), current.Weekday()), len(deploys)).Msg("Deployment frequency")
		current = current.Add(24 * time.Hour)
	}

	return nil
}
