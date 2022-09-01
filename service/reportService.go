package service

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/lib/date"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
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

type ReportService struct {
	deployCommitRelationRepository infra.DeployCommitRelationRepository
	dailyDeploymentCalculator      infra.DailyDeploymentCalculator
	deployCommitRepository         infra.DeployCommitRepository
}

func NewReportService(db *gorm.DB) ReportService {
	deployCommitRelationRepository := infra.DeployCommitRelationRepository{Db: db}
	dailyDeploymentsRepository := infra.DailyDeploymentCalculator{Db: db}
	deployCommitRepository := infra.DeployCommitRepository{Db: db}

	return ReportService{
		deployCommitRelationRepository: deployCommitRelationRepository,
		dailyDeploymentCalculator:      dailyDeploymentsRepository,
		deployCommitRepository:         deployCommitRepository,
	}
}

func (service ReportService) GenerateDeployCalendar(deployMap map[string]int, report infra.ReportGenerator) error {
	today := time.Now()
	current := date.StartOfMonth(today)
	current = current.Add(-24 * time.Hour * time.Duration(current.Weekday()))
	week := []int{}

	markdown := `|Sun|Mon|Tue|Wed|Thu|Fri|Sat|SumOfWeekday|
|---|---|---|---|---|---|---|---|`

	for current.Month() <= today.Month() {
		if current.Weekday() == 0 {
			markdown += "\n|"
		}

		if current.Month() < today.Month() {
			markdown += fmt.Sprintf(" |")
		} else {
			count, ok := deployMap[current.Format("2006-01-02")]
			if current.Weekday() != time.Sunday && current.Weekday() != time.Saturday {
				week = append(week, count)
			}

			if ok {
				markdown += fmt.Sprintf("%v (**%v**)|", current.Day(), count)
			} else {
				markdown += fmt.Sprintf("%v |", current.Day())
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

	report.Append(markdown)

	return nil
}

func (service ReportService) GenerateDeployList(p model.ProjectRepository, ds []model.DeployCommit, report infra.ReportGenerator) error {
	for _, d := range ds {
		commits, err := service.deployCommitRelationRepository.FindByDeployHash(d.Hash)
		if err != nil {
			return err
		}

		report.BulletList(
			[]string{fmt.Sprintf(
				"%v (%v)",
				time.Unix(d.DeployedAt, 0).Format("2006-01-02 15:04:05"),
				markdownCommitLink(p.Org, p.Name, d.Hash),
			)}, 0)

		commitHashes := []string{}
		for _, c := range commits {
			commitHashes = append(commitHashes, markdownCommitLink(p.Org, p.Name, c.CommitHash))
		}
		report.BulletList(commitHashes, 1)
	}

	return nil
}

func (service ReportService) GenerateForRepository(report infra.ReportGenerator, p model.ProjectRepository) error {
	log.Info().Msgf("Generating report for repository %v", p.RepositoryName())

	// Generate a report for this 30 days
	dateCount := 30
	startDate := time.Now().Add(-time.Duration(dateCount) * 24 * time.Hour)
	endDate := time.Now()

	// Calculate deployment frequency and generate the table
	deployments, err := service.dailyDeploymentCalculator.GetDailyDeployment(p.RepositoryName())
	if err != nil {
		return err
	}

	deployMap := map[string]int{}
	for _, d := range deployments {
		deployMap[d.Date] = d.Count
	}

	today := time.Now()
	current := date.StartOfMonth(today)
	for current.Month() <= today.Month() {
		current = current.Add(24 * time.Hour)
	}

	ds, err := service.deployCommitRepository.FindBetweenDeployedAt(p.RepositoryName(), startDate.Unix(), endDate.Unix())
	if err != nil {
		return err
	}

	report.Append(`### Deployment frequency`)
	if err := service.GenerateDeployCalendar(deployMap, report); err != nil {
		return err
	}

	report.Append(fmt.Sprintf(`### Deployments (%v)`, len(ds)))
	if err := service.GenerateDeployList(p, ds, report); err != nil {
		return err
	}

	return nil
}
