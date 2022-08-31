package cmd

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/lib/date"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os/exec"
	"strings"
	"time"
)

func CmdUpdate(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}

	deploymentCommitRepository := infra.DeployCommitRepository{Db: db}

	deploymentCommitRelationRepository := infra.DeployCommitRelationRepository{Db: db}
	if err := deploymentCommitRelationRepository.ResetTable(); err != nil {
		return err
	}

	// update deployment table

	dateCount := 30
	d := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)

	for i := 0; i < dateCount; i++ {
		start, end := date.StartAndEndOfDay(d)

		deploys, err := deploymentCommitRepository.FindBetweenDeployedAt(start.Unix(), end.Unix())
		if err != nil {
			return err
		}

		for _, d := range deploys {
			bin, err := exec.Command("git", "-C", project.Path, "log", "--pretty=format:%H", fmt.Sprintf("%v..%v", d.PreviousHash, d.Hash)).Output()
			if err != nil {
				return err
			}

			relations := []infra.DeployCommitRelation{}
			for _, hash := range strings.Split(string(bin), "\n") {
				if hash == "" {
					continue
				}

				relations = append(relations, infra.DeployCommitRelation{
					DeployHash: d.Hash,
					CommitHash: hash,
				})
			}

			if err := deploymentCommitRelationRepository.Create(relations); err != nil {
				return err
			}

			log.Info().Msgf("%v", string(bin))
		}

		log.Info().Int(fmt.Sprintf("%v (%v)", d.Format("2006-01-02"), d.Weekday()), len(deploys)).Msg("Deployment frequency")
		d = d.Add(24 * time.Hour)
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
