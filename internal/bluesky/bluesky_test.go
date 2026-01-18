package bluesky

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	// 1. Mock the Bluesky PDS server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle 1st request: Create Session
		if r.URL.Path == "/xrpc/com.atproto.server.createSession" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(SessionResponse{
				Did:       "did:plc:12345",
				AccessJwt: "valid.jwt.token",
			})
			return
		}

		// Handle 2nd request: Create Record
		if r.URL.Path == "/xrpc/com.atproto.repo.createRecord" && r.Method == "POST" {
			// Verify Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "Bearer valid.jwt.token" {
				w.WriteHeader(http.StatusOK)
				return // Body doesn't strictly matter for success case in current implementation
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// 2. Test Cases
	tests := []struct {
		name      string
		handle    string
		password  string
		pds       string
		content   string
		expectErr bool
	}{
		{
			name:      "Success",
			handle:    "user.bsky.social",
			password:  "password123",
			pds:       mockServer.URL,
			content:   "Hello Bluesky",
			expectErr: false,
		},
		{
			name:      "Missing Handle",
			handle:    "",
			password:  "password123",
			pds:       mockServer.URL,
			content:   "Hello",
			expectErr: true,
		},
		{
			name:      "Missing Password",
			handle:    "user.bsky.social",
			password:  "",
			pds:       mockServer.URL,
			content:   "Hello",
			expectErr: true,
		},
		{
			name:      "Invalid PDS URL causing network error",
			handle:    "user",
			password:  "pass",
			pds:       "http://invalid-url-that-does-not-exist",
			content:   "Hello",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Post(tt.handle, tt.password, tt.pds, tt.content)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
