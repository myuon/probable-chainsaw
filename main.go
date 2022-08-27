package main

import (
	"github.com/myuon/probable-chainsaw/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	root := &cobra.Command{Use: "keys4"}
	root.AddCommand(&cobra.Command{
		Use:   "init [file]",
		Short: "Initialize keys4 project",
		Run: func(command *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Error().Msg("Specify repositoryUrl")
				return
			}

			if err := cmd.CmdInit("keys4.config.json", args[0], ""); err != nil {
				log.Err(err).Msg("Failed to initialize keys4 project")
				return
			}
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize the project",
		Run: func(command *cobra.Command, args []string) {
			if err := cmd.CmdSync("keys4.config.json"); err != nil {
				log.Err(err).Msg("Failed to synchronize the project")
				return
			}
		},
	})
	if err := root.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute")
	}
}
