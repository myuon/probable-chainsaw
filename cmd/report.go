package cmd

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"sort"
	"time"
)

func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
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

	markdown := `# Report for keys4
`

	// Calculate deployment frequency and generate the table
	type DailyDeployment struct {
		Date  string
		Count int
	}

	deployments := []DailyDeployment{}
	if err := db.
		Model(&model.Deployment{}).
		Group("date(deployed_time)").
		Select("date(deployed_time) as date, count(id) as count").
		Find(&deployments).Error; err != nil {
		return err
	}

	deployMap := map[string]int{}
	for _, d := range deployments {
		deployMap[d.Date] = d.Count
	}

	deployCountMetrics := []int{}

	today := time.Now()
	current := StartOfMonth(today)
	for current.Month() <= today.Month() {
		count := deployMap[current.Format("2006-01")]
		deployCountMetrics = append(deployCountMetrics, count)
		current = current.Add(24 * time.Hour)
	}

	sort.Ints(deployCountMetrics)
	median := deployCountMetrics[len(deployCountMetrics)/2]

	markdown += fmt.Sprintf(`## Deployment frequency
- median: %v

`, median)

	markdown += `|Sun|Mon|Tue|Wed|Thu|Fri|Sat|MedianOfWeekday|
|---|---|---|---|---|---|---|---|`

	current = StartOfMonth(today)
	current = current.Add(-24 * time.Hour * time.Duration(current.Weekday()))
	week := []int{}

	for current.Month() <= today.Month() {
		if current.Weekday() == 0 {
			markdown += "\n|"
		}

		if current.Month() < today.Month() {
			markdown += fmt.Sprintf(" |")
		} else {
			count, ok := deployMap[current.Format("2006-01-02")]
			deployCountMetrics = append(deployCountMetrics, count)
			if current.Weekday() != time.Sunday && current.Weekday() != time.Saturday {
				week = append(week, count)
			}

			if ok {
				markdown += fmt.Sprintf("%v|", count)
			} else {
				markdown += fmt.Sprintf(" |")
			}
		}

		current = current.Add(24 * time.Hour)

		if current.Weekday() == 0 {
			sort.Ints(week)

			markdown += fmt.Sprintf("%v|", week[len(week)/2])
			week = []int{}
		}
	}

	if err := os.WriteFile("report.md", []byte(markdown), 0644); err != nil {
		return err
	}

	log.Info().Msg("✨ Report generated")

	return nil
}
