package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type Repository struct {
	Dir string
}

func (r *Repository) Setup() error {
	dir, err := os.MkdirTemp("", "repository")
	if err != nil {
		return err
	}

	r.Dir = dir

	return nil
}

func (r Repository) CleanUp() error {
	return os.RemoveAll(r.Dir)
}

func (r Repository) Execute() error {
	repo, err := git.PlainClone(r.Dir, false, &git.CloneOptions{
		URL:      "https://github.com/myuon/quartz",
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return err
	}

	if err := commits.ForEach(func(c *object.Commit) error {
		log.Info().Str("created_at", c.Author.When.String()).Msgf("%s", c.Hash)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func run() error {
	repo := Repository{}
	if err := repo.Setup(); err != nil {
		return err
	}
	defer repo.CleanUp()

	if err := repo.Execute(); err != nil {
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
