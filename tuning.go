package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"goTCPTuning/ws"
	"os"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&ws.ClientCmd{}, "websocket")
	subcommands.Register(&ws.ServerCmd{}, "websocket")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
