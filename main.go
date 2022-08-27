package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
}

func (r *Project) Setup() error {
	dir, err := os.MkdirTemp("", "repository")
	if err != nil {
		return err
	}

	r.Dir = dir

	return nil
}

func (r Project) CleanUp() error {
	return os.RemoveAll(r.Dir)
}

func (r Project) FetchCommits() ([]Commit, error) {
	repo, err := git.PlainClone(r.Dir, false, &git.CloneOptions{
		URL:      r.RepositoryUrl,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}

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

func run() error {
	project := Project{
		Dir:           "",
		RepositoryUrl: "https://github.com/myuon/quartz",
	}

	if err := project.Setup(); err != nil {
		return err
	}
	defer project.CleanUp()

	db, err := gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	commits, err := project.FetchCommits()
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&Commit{}); err != nil {
		return err
	}

	if err := db.Create(&commits).Error; err != nil {
		return err
	}

	return nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Started.")

	if err := run(); err != nil {
		log.Err(err)
	}
}
