package main

import (
	"database/sql"
	"fmt"
	"github.com/federicoReghini/gator/internal/cli"
	"github.com/federicoReghini/gator/internal/config"
	"github.com/federicoReghini/gator/internal/database"
	"github.com/federicoReghini/gator/internal/state"
	_ "github.com/lib/pq"
	"os"
)

func main() {

	cmds := cli.NewCommands()

	cmds.Register("login", cli.HandlerLogin)
	cmds.Register("register", cli.Register)
	cmds.Register("reset", cli.Reset)
	cmds.Register("users", cli.Users)
	cmds.Register("agg", cli.Agg)
	cmds.Register("addfeed", cli.MiddlewareLoggedIn(cli.AddFeed))
	cmds.Register("feeds", cli.Feeds)
	cmds.Register("follow", cli.MiddlewareLoggedIn(cli.Follow))
	cmds.Register("following", cli.MiddlewareLoggedIn(cli.Following))
	cmds.Register("unfollow", cli.MiddlewareLoggedIn(cli.Unfollow))
	cmds.Register("browse", cli.Browse)

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
