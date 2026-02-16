package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type SessionResponse struct {
	Did       string `json:"did"`
	AccessJwt string `json:"accessJwt"`
}

type Record struct {
	Type      string    `json:"$type"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateRecordRequest struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     Record `json:"record"`
}

// Post sends a message to Bluesky using the AT Protocol.
func Post(handle, password, pds, content string) error {
	if handle == "" || password == "" {
		return fmt.Errorf("bluesky handle and password are required")
	}

	// 1. Create Session
	sessionURL := fmt.Sprintf("%s/xrpc/com.atproto.server.createSession", pds)
	authBody := map[string]string{
		"identifier": handle,
		"password":   password,
	}
	authJSON, err := json.Marshal(authBody)
	if err != nil {
		return fmt.Errorf("failed to marshal auth body: %w", err)
	}

	resp, err := http.Post(sessionURL, "application/json", bytes.NewBuffer(authJSON)) // #nosec G107
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create session: status code %d", resp.StatusCode)
	}

	var session SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return fmt.Errorf("failed to decode session response: %w", err)
	}

	// 2. Create Record (Post)
	recordURL := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", pds)
	postRecord := Record{
		Type:      "app.bsky.feed.post",
		Text:      content,
		CreatedAt: time.Now().UTC(),
	}
	reqBody := CreateRecordRequest{
		Repo:       session.Did,
		Collection: "app.bsky.feed.post",
		Record:     postRecord,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal record request: %w", err)
	}

	req, err := http.NewRequest("POST", recordURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJwt)

	client := &http.Client{}
	resp, err = client.Do(req) // #nosec G704 -- pds URL is from config, not user input
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Log error response body for debugging
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			log.Errorf("Failed to read error body: %v", err)
		}
		log.Errorf("Bluesky API error: %s", buf.String())
		return fmt.Errorf("failed to create record: status code %d", resp.StatusCode)
	}

	return nil
}
