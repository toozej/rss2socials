package gotify

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/toozej/rss2socials/pkg/config"
)

// Test logFailure function with Gotify notifications
func TestLogFailure(t *testing.T) {
	// Setup a test server to mock Gotify responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set up config for testing
	conf := &config.Config{
		GotifyURL:   server.URL,
		GotifyToken: "test-token",
	}

	logFailure("Test Error", errors.New("this is a test error"), conf)

	// No direct assertions here since the function logs and sends a notification.
	// Verifying that no panic or crash occurs is a basic test for this function.
}

// Test sendGotifyNotification function for success
func TestSendGotifyNotification_Success(t *testing.T) {
	// Setup a test server to mock Gotify responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Query().Get("token") != "test-token" {
			t.Errorf("Expected token 'test-token', got %s", r.URL.Query().Get("token"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := &config.Config{
		GotifyURL:   server.URL,
		GotifyToken: "test-token",
	}

	err := sendGotifyNotification(conf, "Test Title", "Test Message")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test sendGotifyNotification function for failure
func TestSendGotifyNotification_Failure(t *testing.T) {
	// Setup a test server to return an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := &config.Config{
		GotifyURL:   server.URL,
		GotifyToken: "test-token",
	}

	err := sendGotifyNotification(conf, "Test Title", "Test Message")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Test sendGotifyNotification with missing URL
func TestSendGotifyNotification_MissingURL(t *testing.T) {
	conf := &config.Config{
		GotifyURL:   "",
		GotifyToken: "test-token",
	}

	err := sendGotifyNotification(conf, "Test Title", "Test Message")
	if err == nil || err.Error() != "gotify URL or token is not configured" {
		t.Errorf("Expected 'gotify URL or token is not configured', got %v", err)
	}
}

// Test sendGotifyNotification with missing token
func TestSendGotifyNotification_MissingToken(t *testing.T) {
	conf := &config.Config{
		GotifyURL:   "https://example.com",
		GotifyToken: "",
	}

	err := sendGotifyNotification(conf, "Test Title", "Test Message")
	if err == nil || err.Error() != "gotify URL or token is not configured" {
		t.Errorf("Expected 'gotify URL or token is not configured', got %v", err)
	}
}
