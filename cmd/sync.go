package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
)

func CmdSync(configPath string) error {
	project, err := infra.LoadProject(configPath)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}

	commitRepository := infra.CommitRepository{Db: db}
	deployCommitRepository := infra.DeployCommitRepository{Db: db}

	// clear all
	if err := commitRepository.ResetTable(); err != nil {
		return err
	}
	if err := deployCommitRepository.ResetTable(); err != nil {
		return err
	}

	// clone repositories
	for _, p := range project.Repository {
		repo, err := infra.GitOperatorCloneOrPull(p.WorkPath(), fmt.Sprintf("git@github.com:%v/%v.git", p.Org, p.Name))
		if err != nil {
			return err
		}

		// save commits from HEAD
		commits, err := repo.GetCommitsFromHEAD()
		if err != nil {
			return err
		}

		if err := commitRepository.Save(commits); err != nil {
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

		if err := deployCommitRepository.Create(deployCommits); err != nil {
			return err
		}

		if err := commitRepository.UpdateDeployTags("master", commits); err != nil {
			return err
		}
	}

	return nil
}
