package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/lib/date"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os/exec"
	"strings"
	"time"
)

type UpdateService struct {
	commitRepository               infra.CommitRepository
	deployCommitRepository         infra.DeployCommitRepository
	deployCommitRelationRepository infra.DeployCommitRelationRepository
}

func newService(project model.Project) (UpdateService, error) {
	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return UpdateService{}, err
	}

	commitRepository := infra.CommitRepository{Db: db}
	deployCommitRepository := infra.DeployCommitRepository{Db: db}
	deployCommitRelationRepository := infra.DeployCommitRelationRepository{Db: db}

	return UpdateService{
		commitRepository:               commitRepository,
		deployCommitRepository:         deployCommitRepository,
		deployCommitRelationRepository: deployCommitRelationRepository,
	}, nil
}

func (service UpdateService) UpdateRepositoryCommits(p model.ProjectRepository) error {
	repo, err := infra.GitOperatorCloneOrPull(p.WorkPath(), fmt.Sprintf("git@github.com:%v/%v.git", p.Org, p.Name))
	if err != nil {
		return err
	}

	// save commits from HEAD
	commits, err := repo.GetCommitsFromHEAD()
	if err != nil {
		return err
	}

	if err := service.commitRepository.Save(commits); err != nil {
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
			Hash:         c.Hash.String(),
			AuthorName:   c.Author.Name,
			DeployedAt:   c.Author.When.Unix(),
			PreviousHash: previous,
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

func CmdUpdate(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	service, err := newService(project)
	if err != nil {
		return err
	}

	// clear all
	if err := service.commitRepository.ResetTable(); err != nil {
		return err
	}
	if err := service.deployCommitRepository.ResetTable(); err != nil {
		return err
	}

	for _, p := range project.Repository {
		if err := service.UpdateRepositoryCommits(p); err != nil {
			return err
		}
	}

	// update relations
	if err := service.deployCommitRelationRepository.ResetTable(); err != nil {
		return err
	}

	// update deployment table
	for _, p := range project.Repository {

		dateCount := 30
		d := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)

		for i := 0; i < dateCount; i++ {
			start, end := date.StartAndEndOfDay(d)

			deploys, err := service.deployCommitRepository.FindBetweenDeployedAt(start.Unix(), end.Unix())
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

					relations = append(relations, infra.DeployCommitRelation{
						DeployHash: d.Hash,
						CommitHash: hash,
					})
				}

				if err := service.deployCommitRelationRepository.Create(relations); err != nil {
					return err
				}

				log.Info().Msgf("%v", string(bin))
			}

			log.Info().Int(fmt.Sprintf("%v (%v)", d.Format("2006-01-02"), d.Weekday()), len(deploys)).Msg("Deployment frequency")
			d = d.Add(24 * time.Hour)
		}
	}

	type Joined struct {
		DeploymentId model.DeploymentId
		CommitHash   string
		DeployedAt   string
		CommittedAt  int64
	}

	/* Lead time
	if err := db.
		Model(&infra.DeployCommitRelation{}).
		Joins("INNER JOIN deployments ON deployments.id = deploy_commit_relations.deploy_hash").
		Joins("INNER JOIN commits ON deploy_commit_relations.commit_hash = commits.hash").
		Select("deployments.id AS deployment_id, deployments.commit_hash, deployments.deployed_time AS deployed_at, commits.created_at AS committed_at").
		Find(&rs).
		Error; err != nil {
		return err
	}

	relations := []infra.DeployCommitRelation{}
	for _, r := range rs {
		deployedAt, err := time.Parse("2006-01-02 15:04:05", r.DeployedAt)
		if err != nil {
			return err
		}
		committedAt := r.CommittedAt
		_ = deployedAt.Unix() - committedAt

		relations = append(relations, infra.DeployCommitRelation{
			DeployHash: r.CommitHash,
			CommitHash: r.CommitHash,
		})
	}
	if err := deploymentCommitRelationRepository.Save(relations); err != nil {
		return err
	}
	*/

	return nil
}
