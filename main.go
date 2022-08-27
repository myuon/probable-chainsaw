package main

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/cmd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func MarshalStack(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	e, ok := err.(stackTracer)
	if !ok {
		return nil
	}

	msg := err.Error()
	for _, frame := range e.StackTrace() {
		msg += fmt.Sprintf("\n%+s:%d", frame, frame)
	}

	log.Info().Msg(msg)

	return nil
}

func initLogger() {
	zerolog.ErrorStackMarshaler = MarshalStack

	output := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = log.Output(output)
}

func main() {
	initLogger()

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
				log.Err(err).Stack().Msg("Failed to initialize keys4 project")
				return
			}
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize the project",
		Run: func(command *cobra.Command, args []string) {
			if err := cmd.CmdSync("keys4.config.json"); err != nil {
				log.Error().Stack().Err(err).Msg("Failed to synchronize the project")
				return
			}
		},
	})
	if err := root.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute")
		return
	}
}
