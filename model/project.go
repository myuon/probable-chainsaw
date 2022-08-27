package model

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
)

type Project struct {
	Path          string `json:"path"`
	RepositoryUrl string `json:"repositoryUrl"`
	SqliteFile    string `json:"sqliteFile"`
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

func (r Project) FetchCommitsFromBranch(branchName string, repo *git.Repository) (object.CommitIter, error) {
	br, err := repo.Branch(branchName)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Reference(br.Merge, true)
	if err != nil {
		return nil, err
	}

	commits, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, err
	}

	return commits, nil
}
