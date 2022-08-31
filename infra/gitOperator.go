package infra

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
)

type GitOperator struct {
	repo *git.Repository
}

// See also: https://github.com/go-git/go-git/issues/411
func sshAuth() (*ssh.PublicKeys, error) {
	publicKey, err := ssh.NewPublicKeysFromFile("git", fmt.Sprintf("%v/.ssh/id_rsa", os.Getenv("HOME")), "")
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func GitOperatorClone(path string, url string) (GitOperator, error) {
	auth, err := sshAuth()
	if err != nil {
		return GitOperator{}, err
	}

	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return GitOperator{}, err
	}

	return GitOperator{repo: repo}, nil
}

func GitOperatorOpen(path string) (GitOperator, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return GitOperator{}, err
	}

	return GitOperator{repo: repo}, nil
}

// clone or pull
func GitOperatorCloneOrPull(path string, url string) (GitOperator, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Info().Msgf("Repository not exist, cloning %v", url)

		// Create if repository not exist
		return GitOperatorClone(path, url)
	} else {
		log.Info().Msgf("Repository exists, update %v", url)

		// Pull if repository exist
		op, err := GitOperatorOpen(path)
		if err != nil {
			return GitOperator{}, err
		}

		if err := op.repo.Fetch(&git.FetchOptions{
			Force: true,
		}); err != nil {
			if err == git.NoErrAlreadyUpToDate {
				return op, nil
			} else {
				return GitOperator{}, err
			}
		}

		return op, nil
	}
}

func (g GitOperator) GetCommitsFromHEAD() (object.CommitIter, error) {
	commits, err := g.repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (g GitOperator) GetCommitsInBranch(branchName string) (object.CommitIter, error) {
	ref, err := g.repo.ResolveRevision(plumbing.Revision(branchName))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	commits, err := g.repo.Log(&git.LogOptions{
		From: *ref,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return commits, nil
}
