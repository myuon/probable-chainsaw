package cmd

import (
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/model"
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

	// update deployment table
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
	}

	type Joined struct {
		DeploymentId model.DeploymentId
		CommitHash   string
		DeployedAt   string
		CommittedAt  int64
	}

	/* Lead time
	if err := db.
		Model(&infra.DeployCommitRelation{}).
		Joins("INNER JOIN deployments ON deployments.id = deploy_commit_relations.deploy_hash").
		Joins("INNER JOIN commits ON deploy_commit_relations.commit_hash = commits.hash").
		Select("deployments.id AS deployment_id, deployments.commit_hash, deployments.deployed_time AS deployed_at, commits.created_at AS committed_at").
		Find(&rs).
		Error; err != nil {
		return err
	}

	relations := []infra.DeployCommitRelation{}
	for _, r := range rs {
		deployedAt, err := time.Parse("2006-01-02 15:04:05", r.DeployedAt)
		if err != nil {
			return err
		}
		committedAt := r.CommittedAt
		_ = deployedAt.Unix() - committedAt

		relations = append(relations, infra.DeployCommitRelation{
			DeployHash: r.CommitHash,
			CommitHash: r.CommitHash,
		})
	}
	if err := deploymentCommitRelationRepository.Save(relations); err != nil {
		return err
	}
	*/

	return nil
}
