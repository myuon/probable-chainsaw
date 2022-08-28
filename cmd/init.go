package cmd

import (
	"github.com/myuon/probable-chainsaw/infra"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/rs/zerolog/log"
)

func CmdInit(
	configFilePath string,
	repositoryUrl string,
	sqliteFilePath string,
) error {
	if sqliteFilePath == "" {
		sqliteFilePath = "data.db"
	}

	p := model.Project{
		Path:          "",
		RepositoryUrl: repositoryUrl,
		SqliteFile:    sqliteFilePath,
	}

	if err := infra.SaveProject(configFilePath, p); err != nil {
		return err
	}

	log.Info().Msg("Project initialized")

	return nil
}
