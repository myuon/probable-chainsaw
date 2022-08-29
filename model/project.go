package model

import (
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

type Project struct {
	Path          string            `json:"path"`
	RepositoryUrl string            `json:"repositoryUrl"`
	SqliteFile    string            `json:"sqliteFile"`
	Repository    ProjectRepository `json:"repository"`
}

func (r *Project) Setup() error {
	dir, err := os.MkdirTemp("", "repository")
	if err != nil {
		return err
	}

	r.Path = dir

	return nil
}

func (r *Project) CleanUp() error {
	if err := os.RemoveAll(r.Path); err != nil {
		return err
	}
	r.Path = ""

	return nil
}

func (r Project) Clone(auth transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(r.Path, false, &git.CloneOptions{
		URL:      r.RepositoryUrl,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
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
