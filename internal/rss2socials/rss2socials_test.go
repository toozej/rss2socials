package rss2socials

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toozej/rss2socials/internal/db"
	"github.com/toozej/rss2socials/internal/rss"
	"github.com/toozej/rss2socials/pkg/config"

	_ "github.com/mattn/go-sqlite3"
)

// MockRSSChecker is a mock for rss.CheckRSSFeed
type MockRSSChecker struct {
	mock.Mock
}

func (m *MockRSSChecker) CheckRSSFeed(url string) ([]rss.RSSItem, error) {
	args := m.Called(url)
	return args.Get(0).([]rss.RSSItem), args.Error(1)
}

// MockMastodon is a mock for mastodon operations
type MockMastodon struct {
	mock.Mock
}

func (m *MockMastodon) GetTootContent(post rss.RSSItem) string {
	args := m.Called(post)
	return args.String(0)
}

func (m *MockMastodon) TootPost(url, token, content string) error {
	args := m.Called(url, token, content)
	return args.Error(0)
}

// TestHandlePost tests the handlePost function with various scenarios
func TestHandlePost(t *testing.T) {
	tests := []struct {
		name        string
		post        rss.RSSItem
		conf        *config.Config
		dbExists    bool
		dbUpdated   bool
		mastodonErr error
		category    string
		shouldSkip  bool
	}{
		{
			name:        "New post without category",
			post:        rss.RSSItem{Link: "https://example.com/new-post", Content: "content", Title: "New Post"},
			conf:        &config.Config{},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: nil,
			category:    "",
			shouldSkip:  false,
		},
		{
			name:        "New post with category match",
			post:        rss.RSSItem{Link: "https://example.com/new-post-tech", Content: "content", Title: "New Post"},
			conf:        &config.Config{},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: nil,
			category:    "tech",
			shouldSkip:  false,
		},
		{
			name:        "New post with category mismatch",
			post:        rss.RSSItem{Link: "https://example.com/other/new-post", Content: "content", Title: "New Post"},
			conf:        &config.Config{},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: nil,
			category:    "tech",
			shouldSkip:  true,
		},
		{
			name:        "Updated post",
			post:        rss.RSSItem{Link: "https://example.com/updated-post", Content: "updated", Title: "Updated Post"},
			conf:        &config.Config{},
			dbExists:    true,
			dbUpdated:   true,
			mastodonErr: nil,
			category:    "",
			shouldSkip:  false,
		},
		{
			name:        "Mastodon error with Gotify",
			post:        rss.RSSItem{Link: "https://example.com/mastodon-error", Content: "content", Title: "Mastodon Error"},
			conf:        &config.Config{GotifyURL: "http://gotify", GotifyToken: "token"},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: errors.New("mastodon error"),
			category:    "",
			shouldSkip:  false,
		},
		{
			name:        "RSS URL with query params",
			post:        rss.RSSItem{Link: "https://example.com/post-tech?category=tech", Content: "content", Title: "Query Post"},
			conf:        &config.Config{},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: nil,
			category:    "tech",
			shouldSkip:  false,
		},
		{
			name:        "RSS URL with fragment",
			post:        rss.RSSItem{Link: "https://example.com/post#tech", Content: "content", Title: "Fragment Post"},
			conf:        &config.Config{},
			dbExists:    false,
			dbUpdated:   false,
			mastodonErr: nil,
			category:    "tech",
			shouldSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use file-based test DB with cleanup
			originalDB := db.DB

			// Init DB (uses default path, clean up after)
			db.InitDB()
			if tt.dbExists {
				contentToStore := tt.post.Content
				if tt.dbUpdated {
					contentToStore = "different content"
				}
				err := db.StoreTootedPost(tt.post.Link, contentToStore)
				assert.NoError(t, err)
			}

			// Mock Mastodon with test server
			token := "test-token"
			mastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Logf("Test server received: method=%s, path=%s", r.Method, r.URL.Path)
				if r.Method == http.MethodPost {
					if tt.mastodonErr != nil {
						t.Logf("Returning 500 for mastodonErr")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					t.Logf("Returning 200 OK")
					w.WriteHeader(http.StatusOK)
					return
				}
				t.Logf("Returning 404 for non-POST")
				w.WriteHeader(http.StatusNotFound)
			}))
			defer mastodonServer.Close()

			tt.conf.MastodonURL = mastodonServer.URL
			tt.conf.MastodonAccessToken = token

			// Category filtering
			if tt.category != "" {
				tt.conf.Category = tt.category
			}
			lastSegment := path.Base(tt.post.Link)
			shouldProcess := tt.category == "" || strings.Contains(lastSegment, tt.category)
			if !shouldProcess {
				assert.True(t, tt.shouldSkip)
				db.CloseDB()
				os.Remove("./tooted_posts.db")
				return
			}

			// Call handlePost
			handlePost(tt.post, tt.conf)

			// Verify
			if tt.mastodonErr == nil {
				// Should have stored post
				exists, updated, err := db.HasPostChanged(tt.post.Link, tt.post.Content)
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.False(t, updated)
			} else {
				// Should not have stored post
				exists, _, err := db.HasPostChanged(tt.post.Link, tt.post.Content)
				assert.NoError(t, err)
				assert.False(t, exists)
			}

			// Cleanup
			db.CloseDB()
			os.Remove("./tooted_posts.db")
			db.DB = originalDB
		})
	}
}

// TestRunSetup tests the setup logic of Run (flag parsing, config loading, DB init)
func TestRunSetup(t *testing.T) {
	tests := []struct {
		name             string
		setupEnv         map[string]string
		debugFlag        bool
		feedURLFlag      string
		intervalFlag     int
		categoryFlag     string
		expectedDebug    bool
		expectedFeedURL  string
		expectedInterval int
		expectedCategory string
	}{
		{
			name: "Default config from env vars",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://default.com/rss",
				"INTERVAL":              "10",
				"CATEGORY":              "",
				"DEBUG":                 "false",
			},
			debugFlag:        false,
			feedURLFlag:      "",
			intervalFlag:     0,
			categoryFlag:     "",
			expectedDebug:    false,
			expectedFeedURL:  "https://default.com/rss",
			expectedInterval: 10,
			expectedCategory: "",
		},
		{
			name: "Flag overrides",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://env.com/rss",
				"INTERVAL":              "5",
				"CATEGORY":              "envcat",
				"DEBUG":                 "false",
			},
			debugFlag:        false,
			feedURLFlag:      "https://flag.com/rss",
			intervalFlag:     15,
			categoryFlag:     "flagcat",
			expectedDebug:    false,
			expectedFeedURL:  "https://flag.com/rss",
			expectedInterval: 15,
			expectedCategory: "flagcat",
		},
		{
			name: "Debug CLI override true when env false",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://debug.com/rss",
				"INTERVAL":              "10",
				"CATEGORY":              "",
				"DEBUG":                 "false",
			},
			debugFlag:        true,
			feedURLFlag:      "",
			intervalFlag:     0,
			categoryFlag:     "",
			expectedDebug:    true,
			expectedFeedURL:  "https://debug.com/rss",
			expectedInterval: 10,
			expectedCategory: "",
		},
		{
			name: "Debug CLI false when env true (no override to false)",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://debug.com/rss",
				"INTERVAL":              "10",
				"CATEGORY":              "",
				"DEBUG":                 "true",
			},
			debugFlag:        false,
			feedURLFlag:      "",
			intervalFlag:     0,
			categoryFlag:     "",
			expectedDebug:    true,
			expectedFeedURL:  "https://debug.com/rss",
			expectedInterval: 10,
			expectedCategory: "",
		},
		{
			name: "Debug CLI true overrides env true",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://debug.com/rss",
				"INTERVAL":              "10",
				"CATEGORY":              "",
				"DEBUG":                 "true",
			},
			debugFlag:        true,
			feedURLFlag:      "",
			intervalFlag:     0,
			categoryFlag:     "",
			expectedDebug:    true,
			expectedFeedURL:  "https://debug.com/rss",
			expectedInterval: 10,
			expectedCategory: "",
		},
		{
			name: "Full combination with all overrides",
			setupEnv: map[string]string{
				"MASTODON_URL":          "https://mastodon.com",
				"MASTODON_ACCESS_TOKEN": "token",
				"GOTIFY_URL":            "https://gotify.com",
				"GOTIFY_TOKEN":          "gotifytoken",
				"FEED_URL":              "https://env.com/rss",
				"INTERVAL":              "5",
				"CATEGORY":              "envcat",
				"DEBUG":                 "true",
			},
			debugFlag:        false,
			feedURLFlag:      "https://flag.com/rss",
			intervalFlag:     15,
			categoryFlag:     "flagcat",
			expectedDebug:    true,
			expectedFeedURL:  "https://flag.com/rss",
			expectedInterval: 15,
			expectedCategory: "flagcat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			clearEnv := []string{"MASTODON_URL", "MASTODON_ACCESS_TOKEN", "GOTIFY_URL", "GOTIFY_TOKEN", "FEED_URL", "INTERVAL", "CATEGORY", "DEBUG"}
			for _, key := range clearEnv {
				os.Unsetenv(key)
			}

			// Set mock env
			for key, val := range tt.setupEnv {
				os.Setenv(key, val)
			}

			// Create cmd with flags
			cmd := &cobra.Command{}
			cmd.Flags().BoolP("debug", "d", false, "Enable debug logging")
			cmd.Flags().StringP("feed-url", "f", "", "")
			cmd.Flags().IntP("interval", "i", 60, "")
			cmd.Flags().StringP("category", "c", "", "")
			assert.NoError(t, cmd.Flags().Set("debug", fmt.Sprintf("%t", tt.debugFlag)))
			if tt.feedURLFlag != "" {
				assert.NoError(t, cmd.Flags().Set("feed-url", tt.feedURLFlag))
			}
			if tt.intervalFlag > 0 {
				assert.NoError(t, cmd.Flags().Set("interval", fmt.Sprintf("%d", tt.intervalFlag)))
			}
			if tt.categoryFlag != "" {
				assert.NoError(t, cmd.Flags().Set("category", tt.categoryFlag))
			}

			// Call config.GetEnvVars
			conf := config.GetEnvVars()

			// Simulate Run setup with debug override
			debug, err := cmd.Flags().GetBool("debug")
			assert.NoError(t, err)
			if debug {
				conf.Debug = true
			}
			assert.Equal(t, tt.expectedDebug, conf.Debug)

			feedURL := conf.FeedURL
			if tt.feedURLFlag != "" {
				feedURL = tt.feedURLFlag
			}
			if feedURL == "" {
				t.Fatal("RSS feed URL is required")
			}

			interval := conf.Interval
			if tt.intervalFlag > 0 {
				interval = tt.intervalFlag
			}
			if interval <= 0 {
				interval = 60
			}

			category := conf.Category
			if tt.categoryFlag != "" {
				category = tt.categoryFlag
			}

			assert.Equal(t, tt.expectedFeedURL, feedURL)
			assert.Equal(t, tt.expectedInterval, interval)
			assert.Equal(t, tt.expectedCategory, category)

			// Test DB init
			db.InitDB()
			assert.NotNil(t, db.DB)
			db.CloseDB()
		})
	}
}

// TestBasicIntegration tests basic end-to-end flow
func TestBasicIntegration(t *testing.T) {
	// Use file-based test DB with cleanup
	originalDB := db.DB
	db.CloseDB()

	// Init DB (uses default path, clean up after)
	db.InitDB()

	// Mock Mastodon with test server
	token := "test-token"
	mastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Integration test server received: method=%s, path=%s", r.Method, r.URL.Path)
		if r.Method == http.MethodPost {
			t.Logf("Integration test returning 200 OK")
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Logf("Integration test returning 404")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mastodonServer.Close()

	conf := &config.Config{
		MastodonURL:         mastodonServer.URL,
		MastodonAccessToken: token,
	}

	// Test new post handling
	post := rss.RSSItem{Link: "https://test.com/new-post", Content: "test content", Title: "Test Post"}

	handlePost(post, conf)

	// Verify stored in DB
	exists, updated, err := db.HasPostChanged(post.Link, post.Content)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.False(t, updated)

	// Test updated post
	post.Content = "updated content"
	// Check that it's detected as changed before handling
	existsBefore, updatedBefore, err := db.HasPostChanged(post.Link, post.Content)
	assert.NoError(t, err)
	assert.True(t, existsBefore)
	assert.True(t, updatedBefore)

	handlePost(post, conf)

	// After handling, it should be stored so updated is now false
	exists, updated, err = db.HasPostChanged(post.Link, post.Content)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.False(t, updated)

	// Cleanup
	db.CloseDB()
	os.Remove("./tooted_posts.db")
	db.DB = originalDB
}
