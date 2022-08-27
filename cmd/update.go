package cmd

import (
	"fmt"
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
}

func CmdUpdate(configFile string) error {
	project, err := LoadProject(configFile)
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
	date := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)

	for i := 0; i < dateCount; i++ {
		start, end := StartAndEndOfDay(date)

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

		log.Info().Int(fmt.Sprintf("%v (%v)", date.Format("2006-01-02"), date.Weekday()), len(commits)).Msg("Deployment frequency")
		date = date.Add(24 * time.Hour)
	}

	return nil
}
