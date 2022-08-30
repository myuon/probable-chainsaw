package cmd

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/lib/date"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sort"
	"time"
)

func urlForCommit(org string, repositoryName string, commitHash string) string {
	return fmt.Sprintf("https://github.com/%v/%v/commit/%v", org, repositoryName, commitHash)
}

func markdownLink(label string, url string) string {
	return fmt.Sprintf("[%v](%v)", label, url)
}

func markdownCommitLink(org string, repositoryName string, commitHash string) string {
	return markdownLink(commitHash[0:6], urlForCommit(org, repositoryName, commitHash))
}

func CmdReport(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}

	reportGenerator := infra.ReportGenerator{Markdown: ""}

	reportGenerator.Append(`# Report for keys4`)

	deployCommitRelationRepository := infra.DeployCommitRelationRepository{Db: db}

	dailyDeploymentsRepository := infra.DailyDeploymentCalculator{Db: db}

	deployCommitRepository := infra.DeployCommitRepository{Db: db}

	// Generate a report for this 30 days
	dateCount := 30
	startDate := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)
	endDate := time.Now()

	// Calculate deployment frequency and generate the table
	deployments, err := dailyDeploymentsRepository.GetDailyDeployment()
	if err != nil {
		return err
	}

	deployMap := map[string]int{}
	for _, d := range deployments {
		deployMap[d.Date] = d.Count
	}

	deployCountMetrics := []int{}

	today := time.Now()
	current := date.StartOfMonth(today)
	for current.Month() <= today.Month() {
		count := deployMap[current.Format("2006-01")]
		deployCountMetrics = append(deployCountMetrics, count)
		current = current.Add(24 * time.Hour)
	}

	sort.Ints(deployCountMetrics)

	reportGenerator.Append(fmt.Sprintf(`## Repository: %v/%v`, project.Repository.Org, project.Repository.Name))

	reportGenerator.Append(`### Deployment frequency`)

	markdown := `|Sun|Mon|Tue|Wed|Thu|Fri|Sat|SumOfWeekday|
|---|---|---|---|---|---|---|---|`

	current = date.StartOfMonth(today)
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
			w := 0
			for _, c := range week {
				w += c
			}

			markdown += fmt.Sprintf("%v|", w)
			week = []int{}
		}
	}

	reportGenerator.Append(markdown)

	ds, err := deployCommitRepository.FindBetweenDeployedAt(startDate.Unix(), endDate.Unix())
	if err != nil {
		return err
	}

	reportGenerator.Append(fmt.Sprintf(`### Deployments (%v)`, len(ds)))

	for _, d := range ds {
		commits, err := deployCommitRelationRepository.FindByDeployHash(d.Hash)
		if err != nil {
			return err
		}

		reportGenerator.BulletList(
			[]string{fmt.Sprintf(
				"%v (%v)",
				time.Unix(d.DeployedAt, 0).Format("2006-01-02 15:04:05"),
				markdownCommitLink(project.Repository.Org, project.Repository.Name, d.Hash),
			)}, 0)

		commitHashes := []string{}
		for _, c := range commits {
			commitHashes = append(commitHashes, markdownCommitLink(project.Repository.Org, project.Repository.Name, c.CommitHash))
		}
		reportGenerator.BulletList(commitHashes, 1)
	}

	if err := reportGenerator.WriteFile("report.md"); err != nil {
		return err
	}

	log.Info().Msg("✨ Report generated")

	return nil
}
