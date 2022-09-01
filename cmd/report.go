package cmd

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/service"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CmdReport(configFile string, noUpdate bool) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}
	svc := service.NewReportService(db)

	reportGenerator := infra.NewReportGenerator()

	reportGenerator.Append(`# Report for keys4`)
	for _, p := range project.Repository {
		reportGenerator.Append(fmt.Sprintf(`## Repository: %v/%v`, p.Org, p.Name))

		if err := svc.GenerateForRepository(reportGenerator, p); err != nil {
			return err
		}
	}

	if err := reportGenerator.WriteFile("report.md"); err != nil {
		return err
	}

	log.Info().Msg("✨ Report generated")

	return nil
}
