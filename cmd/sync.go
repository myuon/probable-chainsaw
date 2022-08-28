package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/rs/zerolog/log"
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
	project, err := infra.LoadProject(configPath)
	if err != nil {
		return err
	}

	// clone
	if err := project.Setup(); err != nil {
		return err
	}
	defer infra.SaveProject(configPath, project)

	log.Info().Str("path", project.Path).Msg("Setup")

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}

	commitRepository := infra.CommitRepository{
		Db: db,
	}

	// clear all
	if err := commitRepository.ResetTable(); err != nil {
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

	if err := commitRepository.Save(commits); err != nil {
		return err
	}

	// find deployed commits from "deploy" branch
	commits, err = project.FetchCommitsFromBranch("origin/deploy", repo)
	if err != nil {
		return err
	}

	if err := commitRepository.UpdateDeployTags("master", commits); err != nil {
		return err
	}

	return nil
}
