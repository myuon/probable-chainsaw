package cmd

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func StartAndEndOfDay(t time.Time) (time.Time, time.Time) {
	y, m, d := t.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)
	return start, end
}

type DailyDeployment struct {
	Time  time.Time
	Count int
}

func CmdReport(configFile string) error {
	project, err := LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&model.Deployment{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Deployment{}); err != nil {
		return err
	}

	// Deployment frequency: daily

	dateCount := 30
	date := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)

	for i := 0; i < dateCount; i++ {
		start, end := StartAndEndOfDay(date)

		commits := []model.Commit{}
		if err := db.Where("created_at >= ? AND created_at < ? AND deploy_tag != ?", start.Unix(), end.Unix(), "").Find(&commits).Error; err != nil {
			return err
		}

		deployments := []model.Deployment{}
		for _, c := range commits {
			deployments = append(deployments, model.Deployment{
				Id:           model.NewDeploymentId(),
				DeployedTime: time.Unix(c.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			})
		}

		if len(deployments) > 0 {
			if err := db.Save(&deployments).Error; err != nil {
				return err
			}
		}

		log.Info().Int(fmt.Sprintf("%v (%v)", date.Format("2006-01-02"), date.Weekday()), len(commits)).Msg("Deployment frequency")
		date = date.Add(24 * time.Hour)
	}

	return nil
}
