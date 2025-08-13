package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/federicoReghini/Blog-aggregator/internal/cli"
	"github.com/federicoReghini/Blog-aggregator/internal/config"
	"github.com/federicoReghini/Blog-aggregator/internal/database"
	"github.com/federicoReghini/Blog-aggregator/internal/state"
	_ "github.com/lib/pq"
)

func main() {

	cmds := cli.NewCommands()

	cmds.Register("login", cli.HandlerLogin)
	cmds.Register("register", cli.Register)

	if len(os.Args) < 2 {
		fmt.Println("Not enough args, there must be at least 2 args")
		os.Exit(1)
	}

	cfg, _ := config.Read()

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println("db not connect ", err)
	}

	dbQueries := database.New(db)

	appState := &state.State{
		Cfg: cfg,
		Db:  dbQueries,
	}

	// Run Command
	cmdName := os.Args[1]
	args := os.Args[2:]

	cmd := cli.Command{
		Name: cmdName,
		Args: args,
	}

	if err := cmds.Run(appState, cmd); err != nil {
		fmt.Printf("Error running command: %v\n", err)
		os.Exit(1)
	}
}
