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

	// Deployment frequency: daily
	freq := []DailyDeployment{}

	prev := StartOfDay(time.Now().Add(-24 * time.Hour * 7))
	current := StartOfDay(time.Now().Add(-24 * time.Hour * 6))

	for i := 0; i < 7; i++ {
		var commits []model.Commit
		if err := db.Where("created_at >= ? AND created_at < ? AND deploy_tag != ?", prev.Unix(), current.Unix(), "").Find(&commits).Error; err != nil {
			return err
		}

		freq = append(freq, DailyDeployment{
			Time:  prev,
			Count: len(commits),
		})

		prev = prev.Add(24 * time.Hour)
		current = current.Add(24 * time.Hour)
	}

	for _, v := range freq {
		log.Info().Int(fmt.Sprintf("%v(%v)", v.Time.String(), v.Time.Weekday().String()), v.Count).Msg("Deployment frequency")
	}

	return nil
}
