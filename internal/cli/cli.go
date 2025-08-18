package cli

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
func MiddlewareLoggedIn(handler func(s *state.State, cmd Command, user database.User) error) func(*state.State, Command) error {
	return func(s *state.State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			fmt.Println("User not found ", err)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}

func HandlerLogin(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
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
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
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

func scrapeFeeds(s *state.State) error {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	s.Db.MarkFeedFetched(context.Background(), feed.ID)

	f, err := fetchFeed(context.Background(), feed.Url.String)

	for _, item := range f.Channel.Item {
		publishedAt, err := parsePublishedDate(item.PubDate)
		if err != nil {
			log.Printf("failed to parse published date '%s' for post '%s': %v", item.PubDate, item.Title, err)
			continue
		}

		_, err = s.Db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         sql.NullString{String: item.Link, Valid: true},
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: publishedAt,
			FeedID:      uuid.NullUUID{UUID: feed.ID, Valid: true},
		})

		if err != nil {
			// Check if it's a duplicate URL error and ignore it
			if isDuplicateURLError(err) {
				continue
			}
			log.Printf("failed to create post '%s': %v", item.Title, err)
		}
	}

	return nil
}

func parsePublishedDate(pubDate string) (time.Time, error) {
	// Common RSS date formats
	formats := []string{
		time.RFC1123Z,                    // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                     // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,                     // "02 Jan 06 15:04 -0700"
		time.RFC822,                      // "02 Jan 06 15:04 MST"
		"2006-01-02T15:04:05Z07:00",      // ISO 8601
		"2006-01-02 15:04:05",            // Simple format
		"Mon, 2 Jan 2006 15:04:05 -0700", // RFC1123Z without leading zero
		"Mon, 2 Jan 2006 15:04:05 MST",   // RFC1123 without leading zero
	}

	for _, format := range formats {
		if t, err := time.Parse(format, pubDate); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", pubDate)
}

func isDuplicateURLError(err error) bool {
	// This checks for common database unique constraint errors
	errStr := err.Error()
	return contains(errStr, "UNIQUE constraint failed") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "already exists")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func Browse(s *state.State, cmd Command) error {
	limit := int32(2)

	if len(cmd.Args) != 0 {
		// Parse string to int32
		parsedLimit, err := strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid limit value: %s", cmd.Args[0])
		}
		limit = int32(parsedLimit)
	}

	posts, err := s.Db.GetPostsForUser(context.Background(), limit)
	if err != nil {
		return fmt.Errorf("ERROR while getting posts for user: %s", err)

	}
	fmt.Printf("%+v\n", posts)
	return nil

}

func Agg(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {

		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func AddFeed(s *state.State, cmd Command, user database.User) error {

	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       sql.NullString{String: url, Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}
	fmt.Println("Feed created successfully:")
	// Automatically create a feed follow record for the current user
	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Printf("You are now following this feed! (%s)\n", feedFollow.FeedName)

	os.Exit(0)
	return nil
}

func Feeds(s *state.State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())

	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}

	os.Exit(0)
	return nil
}

func Follow(s *state.State, cmd Command, user database.User) error {

	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}

	feed, err := s.Db.GetFeedByUrl(context.Background(), sql.NullString{String: cmd.Args[0], Valid: true})
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	ffRow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %w", err)
	}

	fmt.Println("Feed follow created:")
	printFeedFollow(ffRow.UserName, ffRow.FeedName)
	return nil
}

func Following(s *state.State, cmd Command, user database.User) error {

	feedFollows, err := s.Db.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("couldn't get feed follows: %w", err)
	}

	if len(feedFollows) == 0 {
		fmt.Println("No feed follows found for this user.")
		return nil
	}

	fmt.Printf("Feed follows for user %s:\n", user.Name)
	for _, ff := range feedFollows {
		fmt.Printf("* %s\n", ff.FeedName)
	}

	return nil
}

func printFeedFollow(username, feedname string) {
	fmt.Printf("* User:          %s\n", username)
	fmt.Printf("* Feed:          %s\n", feedname)
}

func Unfollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}

	err := s.Db.DeleteFeedFollowByUserAndFeedUrl(context.Background(), database.DeleteFeedFollowByUserAndFeedUrlParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Url:    sql.NullString{String: cmd.Args[0], Valid: true},
	})

	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}

	fmt.Println("Successfully unfollowed the feed.")
	return nil
}
