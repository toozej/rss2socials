package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnvVars(t *testing.T) {
	tests := []struct {
		name                  string
		setupEnvVars          map[string]string
		setupEnvFileContent   string
		expectedMastodonURL   string
		expectedMastodonToken string
		expectedGotifyURL     string
		expectedGotifyToken   string
		expectedDebug         bool
		expectedFeedURL       string
		expectedInterval      int
	}{
		{
			name: "Valid environment variables",
			setupEnvVars: map[string]string{
				"MASTODON_URL":          "https://mastodon.example.com",
				"MASTODON_ACCESS_TOKEN": "token123",
				"GOTIFY_URL":            "https://gotify.example.com",
				"GOTIFY_TOKEN":          "gotifytoken456",
				"DEBUG":                 "true",
				"FEED_URL":              "https://example.com/rss",
				"INTERVAL":              "30",
			},
			expectedMastodonURL:   "https://mastodon.example.com",
			expectedMastodonToken: "token123",
			expectedGotifyURL:     "https://gotify.example.com",
			expectedGotifyToken:   "gotifytoken456",
			expectedDebug:         true,
			expectedFeedURL:       "https://example.com/rss",
			expectedInterval:      30,
		},
		{
			name:                  "Valid .env file",
			setupEnvFileContent:   "MASTODON_URL=https://mastodon.env.com\nMASTODON_ACCESS_TOKEN=envtoken\nGOTIFY_URL=https://gotify.env.com\nGOTIFY_TOKEN=envgotifytoken\nDEBUG=false\nFEED_URL=https://env.com/rss\nINTERVAL=45\n",
			expectedMastodonURL:   "https://mastodon.env.com",
			expectedMastodonToken: "envtoken",
			expectedGotifyURL:     "https://gotify.env.com",
			expectedGotifyToken:   "envgotifytoken",
			expectedDebug:         false,
			expectedFeedURL:       "https://env.com/rss",
			expectedInterval:      45,
		},
		{
			name: "Environment variable overrides .env file",
			setupEnvVars: map[string]string{
				"MASTODON_URL":          "https://override.com",
				"MASTODON_ACCESS_TOKEN": "token123",
				"GOTIFY_URL":            "https://gotify.example.com",
				"GOTIFY_TOKEN":          "gotifytoken456",
			},
			setupEnvFileContent:   "MASTODON_URL=https://file.com\nMASTODON_ACCESS_TOKEN=filetoken\nGOTIFY_URL=https://gotify.file.com\nGOTIFY_TOKEN=filetoken\n",
			expectedMastodonURL:   "https://override.com",
			expectedMastodonToken: "token123",
			expectedGotifyURL:     "https://gotify.example.com",
			expectedGotifyToken:   "gotifytoken456",
			expectedDebug:         false,
			expectedFeedURL:       "",
			expectedInterval:      60,
		},
		{
			name: "Default interval",
			setupEnvVars: map[string]string{
				"MASTODON_URL":          "https://mastodon.example.com",
				"MASTODON_ACCESS_TOKEN": "token123",
				"GOTIFY_URL":            "https://gotify.example.com",
				"GOTIFY_TOKEN":          "gotifytoken456",
			},
			expectedMastodonURL:   "https://mastodon.example.com",
			expectedMastodonToken: "token123",
			expectedGotifyURL:     "https://gotify.example.com",
			expectedGotifyToken:   "gotifytoken456",
			expectedDebug:         false,
			expectedFeedURL:       "",
			expectedInterval:      60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original directory and change to temp directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}

			tmpDir := t.TempDir()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("Failed to restore original directory: %v", err)
				}
			}()

			// Clear relevant environment variables first
			clearEnvVars := []string{"MASTODON_URL", "MASTODON_ACCESS_TOKEN", "GOTIFY_URL", "GOTIFY_TOKEN", "DEBUG", "FEED_URL", "INTERVAL"}
			for _, key := range clearEnvVars {
				os.Unsetenv(key)
			}

			// Create .env file if applicable
			if tt.setupEnvFileContent != "" {
				envPath := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envPath, []byte(tt.setupEnvFileContent), 0644); err != nil {
					t.Fatalf("Failed to write mock .env file: %v", err)
				}
			}

			// Set mock environment variables (these should override .env file)
			for key, value := range tt.setupEnvVars {
				os.Setenv(key, value)
			}

			// Call function
			conf := GetEnvVars()

			// Verify output
			if conf.MastodonURL != tt.expectedMastodonURL {
				t.Errorf("expected MastodonURL %q, got %q", tt.expectedMastodonURL, conf.MastodonURL)
			}
			if conf.MastodonAccessToken != tt.expectedMastodonToken {
				t.Errorf("expected MastodonAccessToken %q, got %q", tt.expectedMastodonToken, conf.MastodonAccessToken)
			}
			if conf.GotifyURL != tt.expectedGotifyURL {
				t.Errorf("expected GotifyURL %q, got %q", tt.expectedGotifyURL, conf.GotifyURL)
			}
			if conf.GotifyToken != tt.expectedGotifyToken {
				t.Errorf("expected GotifyToken %q, got %q", tt.expectedGotifyToken, conf.GotifyToken)
			}
			if conf.Debug != tt.expectedDebug {
				t.Errorf("expected Debug %v, got %v", tt.expectedDebug, conf.Debug)
			}
			if conf.FeedURL != tt.expectedFeedURL {
				t.Errorf("expected FeedURL %q, got %q", tt.expectedFeedURL, conf.FeedURL)
			}
			if conf.Interval != tt.expectedInterval {
				t.Errorf("expected Interval %d, got %d", tt.expectedInterval, conf.Interval)
			}
		})
	}
}
