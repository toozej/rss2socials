package mastodon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/toozej/rss2socials/internal/rss"
	"github.com/toozej/rss2socials/pkg/config"
)

// Test toot content generation for "Thoughts" posts
func TestGetTootContent_Thoughts(t *testing.T) {
	post := rss.RSSItem{
		Title:   "Thoughts on Go",
		Content: "Go is a great language",
		Link:    "https://example.com/thoughts",
	}

	expected := "Go is a great language - https://example.com/thoughts"
	result := GetTootContent(post, []string{"Thoughts"})

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Test toot content generation for non-"Thoughts" posts
func TestGetTootContent_NewPost(t *testing.T) {
	post := rss.RSSItem{
		Title: "New Blog Post",
		Link:  "https://example.com/blog",
	}

	expected := "New blog post: https://example.com/blog"
	result := GetTootContent(post, nil)

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// MockServer starts a new HTTP test server and returns the server URL along with a function to close the server
func MockServer(statusCode int) (*httptest.Server, string) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	return mockServer, mockServer.URL
}

// Table-driven test for TootPost
func TestTootPost(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "Success",
			statusCode:    http.StatusOK,
			expectedError: false,
		},
		{
			name:          "Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			mockServer, mockServerURL := MockServer(tt.statusCode)
			defer mockServer.Close()

			// Set up config
			conf := config.Config{
				MastodonURL:         mockServerURL,
				MastodonAccessToken: "fake-token",
			}

			// Run the function to test
			err := TootPost(conf.MastodonURL, conf.MastodonAccessToken, "Test toot content")

			// Check if we expect an error or not
			if (err != nil) != tt.expectedError {
				t.Errorf("TestTootPost(%s) failed: expected error: %v, got: %v", tt.name, tt.expectedError, err)
			}
		})
	}
}
