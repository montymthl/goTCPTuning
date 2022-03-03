package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/montymthl/goTCPTuning/ws"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func setupLog(verbose bool, logFile string) {
	var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	if verbose {
		Logger = Logger.Level(zerolog.DebugLevel).With().Caller().Logger()
	}
	if len(logFile) > 0 {
		fp, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Print(err)
			return
		}
		Logger = Logger.Output(fp)
	}
	log.Logger = Logger
}

func main() {
	verbose := flag.Bool("v", false, "Log/Show verbose messages")
	logFile := flag.String("o", "tuning.log", "Log output file")

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&ws.ClientCmd{}, "websocket")
	subcommands.Register(&ws.ServerCmd{}, "websocket")

	flag.Parse()
	setupLog(*verbose, *logFile)
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
