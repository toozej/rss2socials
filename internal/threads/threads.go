package threads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

var threadsAPIURL = "https://graph.threads.net/v1.0"

type CreateContainerResponse struct {
	ID string `json:"id"`
}

type PublishContainerResponse struct {
	ID string `json:"id"`
}

// Post sends a message to Threads using the Threads API.
func Post(userID, token, content string) error {
	if userID == "" || token == "" {
		return fmt.Errorf("threads user ID and token are required")
	}

	// 1. Create a media container
	containerURL := fmt.Sprintf("%s/%s/threads", threadsAPIURL, userID)
	data := url.Values{}
	data.Set("media_type", "TEXT")
	data.Set("text", content)
	data.Set("access_token", token)

	resp, err := http.PostForm(containerURL, data) // #nosec G107
	if err != nil {
		return fmt.Errorf("failed to create threads container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			log.Errorf("Failed to read error body: %v", err)
		}
		log.Errorf("Threads Create Container Error: %s", buf.String())
		return fmt.Errorf("failed to create threads container: status code %d", resp.StatusCode)
	}

	var containerResp CreateContainerResponse
	if err := json.NewDecoder(resp.Body).Decode(&containerResp); err != nil {
		return fmt.Errorf("failed to decode container response: %w", err)
	}

	// 2. Publish the container
	publishURL := fmt.Sprintf("%s/%s/threads_publish", threadsAPIURL, userID)
	publishData := url.Values{}
	publishData.Set("creation_id", containerResp.ID)
	publishData.Set("access_token", token)

	respPublish, err := http.PostForm(publishURL, publishData) // #nosec G107
	if err != nil {
		return fmt.Errorf("failed to publish threads container: %w", err)
	}
	defer respPublish.Body.Close()

	if respPublish.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(respPublish.Body); err != nil {
			log.Errorf("Failed to read error body: %v", err)
		}
		log.Errorf("Threads Publish Error: %s", buf.String())
		return fmt.Errorf("failed to publish threads container: status code %d", respPublish.StatusCode)
	}

	return nil
}
