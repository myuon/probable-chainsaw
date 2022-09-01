package cmd

import (
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/service"
	"github.com/rs/zerolog/log"
	"time"
)

func CmdUpdate(configFile string, start time.Time, end time.Time, targetRepository *string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	svc, err := service.NewUpdateCommitService(project)
	if err != nil {
		return err
	}

	if err := svc.ResetTables(); err != nil {
		return err
	}

	for _, p := range project.Repository {
		if targetRepository != nil {
			if p.Name != *targetRepository && p.RepositoryName() != *targetRepository {
				continue
			}
		}

		if err := svc.UpdateRepositoryCommits(p, project.MainBranch, project.DeployBranch); err != nil {
			return err
		}

		log.Info().Msgf("Updated %v", p.RepositoryName())

		if err := svc.UpdateDeployCommitRelationsOver(p, start, end); err != nil {
			return err
		}

		log.Info().Msgf("Saved commits and deploys of %v", p.RepositoryName())

		if err := svc.UpdateDeployLeadTime(p, start, end); err != nil {
			return err
		}

		log.Info().Msgf("Update lead time of %v", p.RepositoryName())
	}

	return nil
}
