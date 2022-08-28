package infra

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

type CommitRepository struct {
	Db *gorm.DB
}

func (r CommitRepository) ResetTable() error {
	if err := r.Db.Migrator().DropTable(&model.Commit{}); err != nil {
		return err
	}

	if err := r.Db.AutoMigrate(&model.Commit{}); err != nil {
		return err
	}

	return nil
}

func (r CommitRepository) Save(commits object.CommitIter) error {
	if err := commits.ForEach(func(c *object.Commit) error {
		parent := ""
		p, err := c.Parent(0)
		if err == nil {
			parent = p.Hash.String()
		}

		if err := r.Db.Create(&model.Commit{
			Hash:       c.Hash.String(),
			AuthorName: c.Author.Name,
			CreatedAt:  c.Author.When.Unix(),
			DeployTag:  "",
			Parent:     parent,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r CommitRepository) UpdateDeployTags(workingBranch string, commits object.CommitIter) error {
	if err := commits.ForEach(func(c *object.Commit) error {
		// check if this is a merge commit from "master" branch
		if !(strings.Contains(c.Message, "Merge pull request") && strings.Contains(c.Message, workingBranch)) {
			return nil
		}

		// skip if it is not a merge commit
		if c.NumParents() != 2 {
			return nil
		}

		parent, err := c.Parent(1)
		if err != nil {
			return errors.WithStack(err)
		}

		t := model.Commit{}
		if err := r.Db.Where("hash = ?", parent.Hash.String()).First(&t).Error; err != nil {
			return errors.WithStack(err)
		}

		t.DeployTag = c.Hash.String()
		if err := r.Db.Save(&r).Error; err != nil {
			return errors.WithStack(err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
