package cmd

import (
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/myuon/probable-chainsaw/service"
	"time"
)

func CmdUpdate(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	svc, err := service.NewAnalyzeService(project)
	if err != nil {
		return err
	}

	if err := svc.ResetTables(); err != nil {
		return err
	}

	// update deployment table
	datesCount := 28
	for _, p := range project.Repository {
		if err := svc.UpdateRepositoryCommits(p); err != nil {
			return err
		}
		if err := svc.UpdateDeployCommitRelationsOver(p, time.Now().Add(-24*time.Hour*time.Duration(datesCount)), time.Now()); err != nil {
			return err
		}
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
