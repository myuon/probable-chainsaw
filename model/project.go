package model

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pkg/errors"
	"os"
)

type ProjectRepository struct {
	Org  string `json:"org"`
	Name string `json:"name"`
}

func (r ProjectRepository) WorkPath() string {
	return fmt.Sprintf(".work/%v_%v", r.Org, r.Name)
}

func (r ProjectRepository) Clone(auth transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(r.WorkPath(), false, &git.CloneOptions{
		URL:      fmt.Sprintf("git@github.com:%v/%v.git", r.Org, r.Name),
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

func (r Project) FetchCommits(repo *git.Repository) (object.CommitIter, error) {
	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (r Project) FetchCommitsFromBranch(branchName plumbing.Revision, repo *git.Repository) (object.CommitIter, error) {
	ref, err := repo.ResolveRevision(branchName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	commits, err := repo.Log(&git.LogOptions{
		From: *ref,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return commits, nil
}
