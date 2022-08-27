package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/myuon/probable-chainsaw/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

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

	// write commits into database
	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
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

	commits, err := project.FetchCommits(repo)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Commit{}); err != nil {
		return err
	}

	if err := db.Create(&commits).Error; err != nil {
		return err
	}

	return nil
}
