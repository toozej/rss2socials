// Package rss2socials provides the main logic for monitoring RSS feeds and posting updates to Mastodon, Bluesky, and Threads.
// It handles configuration, feed checking, post processing, and integration with other components.
package rss2socials

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/toozej/rss2socials/internal/bluesky"
	"github.com/toozej/rss2socials/internal/db"
	"github.com/toozej/rss2socials/internal/mastodon"
	"github.com/toozej/rss2socials/internal/rss"
	"github.com/toozej/rss2socials/internal/threads"
	"github.com/toozej/rss2socials/pkg/config"
)

func Run(conf config.Config) {
	// Run starts the RSS to Mastodon monitoring loop. It initializes the database,
	// fetches RSS posts at regular intervals, filters them if a category is specified,
	// and handles posting new or updated posts to Mastodon.

	if conf.FeedURL == "" {
		log.Fatal("RSS feed URL is required")
	}

	if conf.Interval <= 0 {
		log.Error("Interval must be a positive integer")
		conf.Interval = 60 // Use default to prevent infinite loop
	}

	db.InitDB() // Initialize SQLite database
	defer db.CloseDB()

	for {
		posts, err := rss.CheckRSSFeed(conf.FeedURL)
		if err != nil {
			log.Printf("Error fetching RSS feed: %v", err)
			continue
		}

		for _, post := range posts {
			if conf.Category != "" {
				// Extract last segment of URL
				lastSegment := path.Base(post.Link)
				if !strings.Contains(lastSegment, conf.Category) {
					log.Debugf("Skipping post %s: category filter '%s' not in URL segment '%s'", post.Title, conf.Category, lastSegment)
					continue
				}
			}
			handlePost(post, &conf)
		}

		// Sleep for the configured interval before checking again
		time.Sleep(time.Duration(conf.Interval) * time.Minute)
	}
}

func handlePost(post rss.RSSItem, conf *config.Config) {
	// handlePost processes an RSS item, checks if it needs to be posted or updated on Mastodon,
	// sends the toot if necessary, and stores the post in the database.
	exists, updated, err := db.HasPostChanged(post.Link, post.Content)
	if err != nil {
		log.Error("Database error: ", err)
		return
	}

	var tootContent string
	var isUpdate bool

	switch {
	case exists && updated:
		// Post exists but is updated
		log.Printf("Post has been updated: %s", post.Title)
		tootContent = fmt.Sprintf("Blog post has been updated: %s", post.Link)
		isUpdate = true
	case !exists:
		// New post
		tootContent = mastodon.GetTootContent(post, conf.SkipPrefixCategories)
		isUpdate = false
	default:
		// Post exists but unchanged
		return
	}

	err = mastodon.TootPost(conf.MastodonURL, conf.MastodonAccessToken, tootContent)
	if err != nil {
		if isUpdate {
			log.Error("Failed to toot updated post: ", err)
		} else {
			log.Printf("Failed to toot new post: %v", err)
		}
		return
	}

	// Store the current content after successful toot
	err = db.StoreTootedPost(post.Link, post.Content)
	if err != nil {
		log.Error("Storing post toot in database failed: ", err)
	}

	// Post to Bluesky (Fire and forget or log error, but don't block DB update which is primary?
	// Actually we should post to all and then mark as done. Existing logic marks as done after Mastodon.
	// For now, I will add Bluesky posting here. Ideally we should have a 'posted_to' table but scope is refactor.)
	if conf.BlueskyHandle != "" && conf.BlueskyPassword != "" {
		log.Infof("Posting to Bluesky: %s", post.Title)
		if err := bluesky.Post(conf.BlueskyHandle, conf.BlueskyPassword, conf.BlueskyPDS, tootContent); err != nil {
			log.Errorf("Failed to post to Bluesky: %v", err)
		} else {
			log.Info("Successfully posted to Bluesky")
		}
	}

	// Post to Threads
	if conf.ThreadsUserID != "" && conf.ThreadsToken != "" {
		log.Infof("Posting to Threads: %s", post.Title)
		if err := threads.Post(conf.ThreadsUserID, conf.ThreadsToken, tootContent); err != nil {
			log.Errorf("Failed to post to Threads: %v", err)
		} else {
			log.Info("Successfully posted to Threads")
		}
	}
}
