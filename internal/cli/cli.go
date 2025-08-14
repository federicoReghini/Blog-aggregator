package cli

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
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

func Reset(s *state.State, cmd Command) error {

	if err := s.Db.ResetUsers(context.Background()); err != nil {
		fmt.Println("Something went wrong while deleting whole user records")
		os.Exit(1)
	}

	fmt.Println("User records deleted successfully")
	os.Exit(0)
	return nil
}

func Users(s *state.State, cmd Command) error {
	if users, err := s.Db.GetUsers(context.Background()); err == nil {

		for _, user := range users {
			if user.Name == s.Cfg.CurrentUserName {
				fmt.Printf("%s (current)", user.Name)
			} else {
				fmt.Println(user.Name)
			}
		}

		os.Exit(0)
	}
	fmt.Println("Something went wrong while getting all the users")
	os.Exit(1)
	return nil
}

// RSS PART
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {
		return nil, errors.New("Something went wrong while prepering request")
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return nil, errors.New("Something went wrong while prepering request")
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)

	}

	return &feed, nil
}

func Agg(s *state.State, cmd Command) error {

	url := "https://www.wagslane.dev/index.xml"

	if len(cmd.Args) != 0 {
		url = cmd.Args[0]
	}

	feed, err := fetchFeed(context.Background(), url)

	if err != nil {
		os.Exit(1)
		return err
	}

	for i, item := range feed.Channel.Item {
		fmt.Printf("Feed %d: %+v\n", i, item)
	}

	return nil
}
