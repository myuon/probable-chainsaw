package cmd

import (
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/service"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CmdReport(configFile string) error {
	project, err := infra.LoadProject(configFile)
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(project.SqliteFile), &gorm.Config{})
	if err != nil {
		return err
	}
	svc := service.NewReportService(db)

	reportGenerator := infra.ReportGenerator{Markdown: ""}

	reportGenerator.Append(`# Report for keys4`)
	for _, p := range project.Repository {
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
