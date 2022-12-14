package main

import (
	"fmt"
	"github.com/myuon/probable-chainsaw/cmd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"time"
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

type ReportOptions struct {
	NoUpdate bool
	Month    int
}

func (r ReportOptions) Span() (time.Time, time.Time) {
	y, _, d := time.Now().Date()
	start := time.Date(y, time.Month(r.Month), d, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	return start, end
}

func main() {
	initLogger()

	configFile := "keys4.config.yml"
	options := ReportOptions{}

	root := &cobra.Command{Use: "keys4"}
	root.AddCommand(&cobra.Command{
		Use:   "init [file]",
		Short: "Initialize keys4 project",
		Run: func(command *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Error().Msg("Specify repositoryUrl")
				return
			}

			if err := cmd.CmdInit(configFile, args[0], ""); err != nil {
				log.Err(err).Stack().Msg("Failed to initialize keys4 project")
				return
			}
		},
	})

	reportCmd := cobra.Command{
		Use:   "report",
		Short: "Report the project",
		Run: func(command *cobra.Command, args []string) {
			start, end := options.Span()

			if !options.NoUpdate {
				if err := cmd.CmdUpdate(configFile, start, end, nil); err != nil {
					log.Error().Stack().Err(err).Msg("Failed to report the project")
					return
				}
			}

			if err := cmd.CmdReport(configFile, start, end); err != nil {
				log.Error().Stack().Err(err).Msg("Failed to report the project")
				return
			}
		},
	}
	root.AddCommand(&reportCmd)
	reportCmd.Flags().BoolVar(&options.NoUpdate, "noupdate", false, "Do not update the project")
	reportCmd.Flags().IntVar(&options.Month, "month", int(time.Now().Month()), "Month to report")

	// loe-level commands
	root.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update the data and statistics",
		Run: func(command *cobra.Command, args []string) {
			start, end := options.Span()

			var targetRepo *string
			if len(args) > 0 {
				targetRepo = &args[0]
			}

			if err := cmd.CmdUpdate(configFile, start, end, targetRepo); err != nil {
				log.Error().Stack().Err(err).Msg("Failed to report the project")
				return
			}
		},
	})
	if err := root.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute")
		return
	}
}
