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

func (service ReportService) GenerateDeployCalendar(deployMap map[string]int, report infra.ReportGenerator, start time.Time, end time.Time) error {
	current := start
	current = current.Add(-24 * time.Hour * time.Duration(current.Weekday()))
	week := []int{}

	markdown := `|Sun|Mon|Tue|Wed|Thu|Fri|Sat|SumOfWeekday|
|---|---|---|---|---|---|---|---|`

	for current.Before(end) {
		if current.Weekday() == 0 {
			markdown += "\n|"
		}

		if current.Month() < start.Month() {
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

		current = current.AddDate(0, 0, 1)

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
	list := []string{}
	for _, d := range ds {
		commits, err := service.deployCommitRelationRepository.FindByDeployHash(d.Hash)
		if err != nil {
			return err
		}

		list = append(list, fmt.Sprintf(
			"%v (%v), Lead Time: %v, %v commits",
			time.Unix(d.DeployedAt, 0).Format("2006-01-02 15:04:05"),
			markdownCommitLink(p.Org, p.Name, d.Hash),
			date.SecondsInHumanReadableFormat(d.LeadTime),
			len(commits),
		))
	}

	report.BulletList(list, 0)

	return nil
}

func (service ReportService) GenerateContributors(contributorsMap map[string]model.DeployCommits, report infra.ReportGenerator) error {
	type pair struct {
		author string
		len    int
	}
	pairs := []pair{}
	for k, v := range contributorsMap {
		pairs = append(pairs, pair{
			author: k,
			len:    len(v),
		})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].len > pairs[j].len
	})

	list := []string{}
	for i, p := range pairs {
		if i >= 3 {
			break
		}

		icon := ""
		if i == 0 {
			icon = "????"
		}

		list = append(list, fmt.Sprintf("%v **%v** (%v commits)", icon, p.author, p.len))
	}

	report.BulletList(list, 0)

	return nil
}

func (service ReportService) GenerateForRepository(report infra.ReportGenerator, p model.ProjectRepository, start time.Time, end time.Time) error {
	log.Info().Msgf("Generating report for repository %v", p.RepositoryName())

	ds, err := service.deployCommitRepository.FindBetweenDeployedAt(p.RepositoryName(), start.Unix(), end.Unix())
	if err != nil {
		return err
	}

	report.Append(`### Summary`)
	report.BulletList([]string{
		fmt.Sprintf("Total Deployments: %v", len(ds)),
		fmt.Sprintf("Avg Lead Time: %v", date.SecondsInHumanReadableFormat(ds.LeadTimeAvg())),
	}, 0)

	report.Append(`### Deployment frequency`)
	if err := service.GenerateDeployCalendar(ds.DeployDailyMap(), report, start, end); err != nil {
		return err
	}

	report.Append(fmt.Sprintf(`### Deployments (%v)`, len(ds)))
	if err := service.GenerateDeployList(p, ds, report); err != nil {
		return err
	}

	return nil
}

func (service ReportService) GenerateTotal(report infra.ReportGenerator, start time.Time, end time.Time) error {
	log.Info().Msgf("Generating report for total")

	ds, err := service.deployCommitRepository.FindBetweenDeployedAtAnyRepository(start.Unix(), end.Unix())
	if err != nil {
		return err
	}

	report.Append(`### Summary`)
	report.BulletList([]string{
		fmt.Sprintf("Total Deployments: %v", len(ds)),
		fmt.Sprintf("Avg Lead Time: %v", date.SecondsInHumanReadableFormat(ds.LeadTimeAvg())),
	}, 0)

	report.Append(`### Deployment frequency`)
	if err := service.GenerateDeployCalendar(ds.DeployDailyMap(), report, start, end); err != nil {
		return err
	}

	return nil
}
