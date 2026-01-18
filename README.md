# rss2socials

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/toozej/rss2socials)
[![Go Report Card](https://goreportcard.com/badge/github.com/toozej/rss2socials)](https://goreportcard.com/report/github.com/toozej/rss2socials)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/toozej/rss2socials/cicd.yaml)
![Docker Pulls](https://img.shields.io/docker/pulls/toozej/rss2socials)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/toozej/rss2socials/total)

rss2socials is a CLI tool that monitors an RSS feed for new posts and automatically posts updates to specified social platforms (Mastodon, Bluesky, Threads). This application is designed for easy configuration and seamless integration.

## Features
- Periodically checks an RSS feed for new or updated posts.
- Posts updates to configured social platforms (Mastodon, Bluesky, Threads).
- Stores previously posted items in an SQLite database to avoid duplicates.
- Configurable check interval and customizable content.
- Debug mode for detailed logging.

## Installation
### Prerequisites
- Go (version 1.25 or later)
- SQLite for database management
- Make

### Steps
1.	Clone the repository:
```bash
git clone https://github.com/toozej/rss2socials.git
cd rss2socials
```

2.	Build the executable:
`make build`

## Usage
1.	Set Environment Variables:
    Create a .env file in the root of your project or set the required environment variables directly:

    ```
    # Mastodon
    MASTODON_URL=https://your-mastodon-instance
    MASTODON_TOKEN=your-access-token
    
    # Bluesky
    BLUESKY_HANDLE=your.handle.bsky.social
    BLUESKY_PASSWORD=your-app-password
    
    # Threads
    THREADS_USER_ID=your-user-id
    THREADS_TOKEN=your-access-token
    
    # General
    FEED_URL=https://example.com/rss
    ```

    Alternatively, you can provide parameters as command-line flags.

2.	Run the application:
    ```bash
    ./rss2socials --feed-url "https://example.com/rss" --interval 60
    ```

    `--feed-url`: The URL of the RSS feed to monitor.
    `--interval`: The interval in minutes for checking the RSS feed (default is 60 minutes).

3. Enable Debug Mode:
Use the --debug flag to enable debug-level logging for troubleshooting.
```bash
./rss2socials --debug
```


## Major Components
### Command Structure (cmd/rss2socials/root.go)
- Defines the main rss2socials command and its subcommands (man and version).
- Sets up CLI flags and binds them to configuration via Viper.

### Configuration (pkg/config/config.go)
- Loads configuration from environment variables and the .env file if present.

### RSS Handling (internal/rss/rss.go)
- Fetches and parses the RSS feed.
- Provides hashing functionality to detect changes in post content.

### Social Integrations
- **Mastodon**: `internal/mastodon`
- **Bluesky**: `internal/bluesky`
- **Threads**: `internal/threads`

### Database Management (internal/db/db.go)
- Manages an SQLite database to store and check previously posted items.

## update golang version
- `make update-golang-version`

