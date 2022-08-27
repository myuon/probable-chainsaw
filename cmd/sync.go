package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

// See also: https://github.com/go-git/go-git/issues/411
func SshAuth() (*ssh.PublicKeys, error) {
	publicKey, err := ssh.NewPublicKeysFromFile("git", fmt.Sprintf("%v/.ssh/id_rsa", os.Getenv("HOME")), "")
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func CmdSync(configPath string) error {
	project, err := LoadProject(configPath)
	if err != nil {
		return err
	}

	// clone
	if err := project.Setup(); err != nil {
		return err
	}
	defer project.CleanUp()

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}

	// clear all
	if err := db.Migrator().DropTable(&model.Commit{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Commit{}); err != nil {
		return err
	}

	sshAuth, err := SshAuth()
	if err != nil {
		return err
	}
	repo, err := project.Clone(sshAuth)
	if err != nil {
		return err
	}

	// save commits from HEAD
	commits, err := project.FetchCommits(repo)
	if err != nil {
		return err
	}
	if err := commits.ForEach(func(c *object.Commit) error {
		if err := db.Create(&model.Commit{
			Hash:       c.Hash.String(),
			AuthorName: c.Author.Name,
			CreatedAt:  c.Author.When.Unix(),
			DeployTag:  "",
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// find deployed commits from "deploy" branch
	commits, err = project.FetchCommitsFromBranch("deploy", repo)
	if err != nil {
		return err
	}

	if err := commits.ForEach(func(c *object.Commit) error {
		// skip if it is not a merge commit
		if c.NumParents() != 2 {
			return nil
		}

		parent, err := c.Parent(1)
		if err != nil {
			return err
		}

		r := model.Commit{}
		if err := db.Where("hash = ?", parent.Hash).Find(&r).Error; err != nil {
			return err
		}

		r.DeployTag = c.Hash.String()
		if err := db.Save(&r).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
