package threads

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	// 1. Mock the Threads API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle 1st request: Create Container
		if r.URL.Path == "/123456/threads" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CreateContainerResponse{ID: "container_id_123"})
			return
		}

		// Handle 2nd request: Publish Container
		if r.URL.Path == "/123456/threads_publish" && r.Method == "POST" {
			// Verify creation_id parameter
			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.Form.Get("creation_id") == "container_id_123" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(PublishContainerResponse{ID: "publish_id_456"})
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// 2. Override the API URL for testing using the unexported variable (allowed in same package tests)
	originalURL := threadsAPIURL
	threadsAPIURL = mockServer.URL
	defer func() { threadsAPIURL = originalURL }()

	// 3. Test Cases
	tests := []struct {
		name      string
		userID    string
		token     string
		content   string
		expectErr bool
	}{
		{
			name:      "Success",
			userID:    "123456",
			token:     "valid_token",
			content:   "Hello Threads",
			expectErr: false,
		},
		{
			name:      "Missing UserID",
			userID:    "",
			token:     "valid_token",
			content:   "Hello",
			expectErr: true,
		},
		{
			name:      "Missing Token",
			userID:    "123456",
			token:     "",
			content:   "Hello",
			expectErr: true,
		},
		{
			name:      "API Failure (Invalid User)",
			userID:    "invalid_user",
			token:     "valid_token",
			content:   "Hello",
			expectErr: true, // Mock server returns 404 for paths not starting with /123456/
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Post(tt.userID, tt.token, tt.content)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
