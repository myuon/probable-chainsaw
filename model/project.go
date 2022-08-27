package model

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
)

type Commit struct {
	Hash       string `sql:"primaryKey"`
	AuthorName string
	CreatedAt  int64
}

type Project struct {
	Dir           string `json:"path"`
	RepositoryUrl string `json:"repositoryUrl"`
	SqliteFile    string `json:"sqliteFile"`
}

func (r *Project) Setup() error {
	dir, err := os.MkdirTemp("", "repository")
	if err != nil {
		return err
	}

	r.Dir = dir

	return nil
}

func (r *Project) CleanUp() error {
	if err := os.RemoveAll(r.Dir); err != nil {
		return err
	}
	r.Dir = ""

	return nil
}

func (r Project) Clone(auth transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(r.Dir, false, &git.CloneOptions{
		URL:      r.RepositoryUrl,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r Project) FetchCommits(repo *git.Repository) ([]Commit, error) {
	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	result := []Commit{}
	if err := commits.ForEach(func(c *object.Commit) error {
		result = append(result, Commit{
			Hash:       c.Hash.String(),
			AuthorName: c.Author.Name,
			CreatedAt:  c.Author.When.Unix(),
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}
