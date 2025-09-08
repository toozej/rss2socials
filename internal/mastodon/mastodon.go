// Package mastodon provides functionality for interacting with the Mastodon API.
// It includes utilities for formatting toot content from RSS items and sending posts to Mastodon instances.
package mastodon

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/toozej/rss2socials/internal/rss"
)

// GetTootContent constructs the toot message depending on the post title
func GetTootContent(post rss.RSSItem, skipPrefixCategories []string) string {
	// GetTootContent formats the RSS item into a Mastodon toot message.
	// It customizes the content based on the post title, using the skipPrefixCategories list.
	for _, cat := range skipPrefixCategories {
		if strings.HasPrefix(post.Title, cat) {
			return fmt.Sprintf("%s - %s", post.Content, post.Link)
		}
	}
	return fmt.Sprintf("New blog post: %s", post.Link)
}

// TootPost sends a post to Mastodon
func TootPost(mastodonURL, mastodonToken, content string) error {
	// TootPost sends a toot to the specified Mastodon instance using the provided access token.
	// It constructs an HTTP POST request to the Mastodon API and handles the response.
	if mastodonURL == "" || mastodonToken == "" {
		return fmt.Errorf("mastodon URL and token must be set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	formData := fmt.Sprintf("status=%s", content)
	req, err := http.NewRequest("POST", mastodonURL+"/api/v1/statuses", strings.NewReader(formData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mastodonToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	return nil
}
