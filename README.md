# Gator cli

## Prerequisites

You'll need to have the following installed on your system:

1. **PostgreSQL** - The program uses PostgreSQL as its database
2. **Go** - Required to build and install the CLI tool

## Installation

Install the gator CLI using Go:

```bash
go install github.com/your-username/gator@latest
```

Or if you're working with the local source:

```bash
go install .
```

## Configuration Setup

1. **Database Setup**: Make sure PostgreSQL is running and create a database for gator.

2. **Config File**: The program expects a `.gatorconfig.json` file. Based on the code structure, this should contain your database connection details:

```json
{
  "db_url": "postgres://username:password@localhost/gator_db?sslmode=disable"
}
```

Place this file in your home directory or the directory where you'll run gator commands.

## Available Commands

Once set up, you can use several commands:

- **User Management**:
  - `gator register <username>` - Register a new user
  - `gator login <username>` - Login as a user
  - `gator users` - List all users

- **Feed Management**:
  - `gator addfeed <name> <url>` - Add an RSS feed
  - `gator feeds` - List all feeds
  - `gator follow <feed_name>` - Follow a feed
  - `gator unfollow <feed_name>` - Unfollow a feed
  - `gator following` - Show feeds you're following

- **Content**:
  - `gator browse <limit>` - Browse recent posts
  - `gator agg` - Aggregate/fetch new posts from feeds

- **Other**:
  - `gator reset` - Reset the database

The CLI will automatically handle database migrations and setup when you first run it.
