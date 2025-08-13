package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/federicoReghini/Blog-aggregator/internal/database"
	"github.com/federicoReghini/Blog-aggregator/internal/state"
	"github.com/google/uuid"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	handlers map[string]func(*state.State, Command) error
}

func (c *Commands) Run(s *state.State, cmd Command) error {

	if callback, exist := c.handlers[cmd.Name]; exist {
		return callback(s, cmd)
	}

	return errors.New("Command not found")

}

func (c *Commands) Register(name string, callback func(*state.State, Command) error) {
	c.handlers[name] = callback
}

func NewCommands() *Commands {
	return &Commands{
		handlers: make(map[string]func(*state.State, Command) error),
	}
}

func cleanString(s string) []string {
	if s == "" {
		return nil
	}
	// Remove leading and trailing whitespace
	s = strings.TrimSpace(s)
	// Replace multiple spaces with a single space
	s = strings.Join(strings.Fields(s), " ")

	//Lowercase the string
	s = strings.ToLower(s)

	prompt := strings.Fields(s)
	return prompt
}

func GetCmdNameAndArgs(prompt string) (string, []string) {

	input := cleanString(prompt)

	cmdName := input[0]
	args := input[1:]

	return cmdName, args

}

func HandlerLogin(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Username is required")
	}

	if _, err := s.Db.GetUser(context.Background(), cmd.Args[0]); err != nil {
		fmt.Println("User not found, register before log in")
		os.Exit(1)

	}

	s.Cfg.SetUser(cmd.Args[0])

	fmt.Printf("User %s has beeen set", cmd.Args[0])

	return nil
}

func Register(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("A name is required")
	}

	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        int32(uuid.New()[0]),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	})

	if err != nil {
		os.Exit(1)
	}

	s.Cfg.SetUser(user.Name)

	fmt.Printf("User registered successfully %+v\n", user)

	return nil
}
