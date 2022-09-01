package model

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
)

type ProjectRepository struct {
	Org  string `json:"org"`
	Name string `json:"name"`
}

func (r ProjectRepository) WorkPath() string {
	return fmt.Sprintf(".work/%v_%v", r.Org, r.Name)
}

func (r ProjectRepository) RepositoryName() string {
	return fmt.Sprintf("%v/%v", r.Org, r.Name)
}

func (r ProjectRepository) GitHubUrl() string {
	return fmt.Sprintf("git@github.com:%v/%v.git", r.Org, r.Name)
}

func (r ProjectRepository) Clone(auth transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(r.WorkPath(), false, &git.CloneOptions{
		URL:      r.GitHubUrl(),
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

type Project struct {
	RepositoryUrl string              `json:"repositoryUrl"`
	SqliteFile    string              `json:"sqliteFile"`
	Repository    []ProjectRepository `json:"repository"`
}
