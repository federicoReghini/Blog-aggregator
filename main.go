package main

import (
	"fmt"
	"github.com/federicoReghini/Blog-aggregator/internal/cli"
	"github.com/federicoReghini/Blog-aggregator/internal/config"
	"github.com/federicoReghini/Blog-aggregator/internal/state"
	"os"
)

func main() {

	cmds := cli.NewCommands()

	cmds.Register("login", cli.HandlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("Not enough args, there must be at least 2 args")
		os.Exit(1)
	}

	cfg, _ := config.Read()

	appState := &state.State{
		Cfg: cfg,
	}

	// Run Command
	cmdName := os.Args[1]
	args := os.Args[2:]

	cmd := cli.Command{
		Name: cmdName,
		Args: args,
	}

	err := cmds.Run(appState, cmd)
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		os.Exit(1)
	}
}
