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

type DeploymentCommitRelation struct {
	DeploymentId model.DeploymentId `gorm:"primaryKey"`
	CommitHash   string             `gorm:"primaryKey"`
	LeadTime     int64
}

func CmdUpdate(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&model.Deployment{}, &DeploymentCommitRelation{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Deployment{}, &DeploymentCommitRelation{}); err != nil {
		return err
	}

	// update deployment table

	dateCount := 30
	d := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)

	for i := 0; i < dateCount; i++ {
		start, end := date.StartAndEndOfDay(d)

		commits := []model.Commit{}
		if err := db.Where("created_at >= ? AND created_at < ? AND deploy_tag != ?", start.Unix(), end.Unix(), "").Find(&commits).Error; err != nil {
			return err
		}

		for _, c := range commits {
			deployment := model.Deployment{
				Id:           model.NewDeploymentId(),
				DeployedTime: time.Unix(c.CreatedAt, 0).Format("2006-01-02 15:04:05"),
				CommitHash:   c.Hash,
			}

			if err := db.Create(&deployment).Error; err != nil {
				return err
			}

			bin, err := exec.Command("git", "-C", project.Path, "log", "--pretty=format:%H", fmt.Sprintf("%v..%v", c.Parent, c.Hash)).Output()
			if err != nil {
				return err
			}

			relations := []DeploymentCommitRelation{}
			for _, hash := range strings.Split(string(bin), "\n") {
				if hash == "" {
					continue
				}

				relations = append(relations, DeploymentCommitRelation{
					DeploymentId: deployment.Id,
					CommitHash:   hash,
				})
			}

			if err := db.Create(&relations).Error; err != nil {
				return err
			}

			log.Log().Msgf("%v", string(bin))
		}

		log.Info().Int(fmt.Sprintf("%v (%v)", d.Format("2006-01-02"), d.Weekday()), len(commits)).Msg("Deployment frequency")
		d = d.Add(24 * time.Hour)
	}

	type Joined struct {
		DeploymentId model.DeploymentId
		CommitHash   string
		DeployedAt   string
		CommittedAt  int64
	}

	rs := []Joined{}

	if err := db.
		Model(&DeploymentCommitRelation{}).
		Joins("INNER JOIN deployments ON deployments.id = deployment_commit_relations.deployment_id").
		Joins("INNER JOIN commits ON deployment_commit_relations.commit_hash = commits.hash").
		Select("deployments.id AS deployment_id, deployments.commit_hash, deployments.deployed_time AS deployed_at, commits.created_at AS committed_at").
		Find(&rs).
		Error; err != nil {
		return err
	}

	for _, r := range rs {
		deployedAt, err := time.Parse("2006-01-02 15:04:05", r.DeployedAt)
		if err != nil {
			return err
		}
		committedAt := r.CommittedAt

		if err := db.Save(&DeploymentCommitRelation{
			DeploymentId: r.DeploymentId,
			CommitHash:   r.CommitHash,
			LeadTime:     deployedAt.Unix() - committedAt,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
