package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Table-driven tests for CheckRSSFeed with various scenarios
func TestCheckRSSFeed(t *testing.T) {
	tests := []struct {
		name          string
		xmlContent    string
		statusCode    int
		expectedPosts int
		expectedError bool
		expectedTitle string
	}{
		{
			name: "Valid RSS feed",
			xmlContent: `
				<rss>
					<channel>
						<title>Test Blog</title>
						<item>
							<title>Test Post</title>
							<link>https://example.com/test-post</link>
							<description>This is a test post</description>
						</item>
						<item>
							<title>Second Post</title>
							<link>https://example.com/second-post</link>
							<description>Second test post</description>
						</item>
					</channel>
				</rss>`,
			statusCode:    200,
			expectedPosts: 2,
			expectedError: false,
			expectedTitle: "Test Post",
		},
		{
			name: "Empty RSS feed",
			xmlContent: `
				<rss>
					<channel>
						<title>Empty Blog</title>
					</channel>
				</rss>`,
			statusCode:    200,
			expectedPosts: 0,
			expectedError: false,
		},
		{
			name:          "Invalid XML",
			xmlContent:    `Invalid XML content`,
			statusCode:    200,
			expectedPosts: 0,
			expectedError: true,
		},
		{
			name:          "HTTP error 404",
			xmlContent:    ``,
			statusCode:    404,
			expectedPosts: 0,
			expectedError: true,
		},
		{
			name: "RSS with different structure",
			xmlContent: `
				<rss>
					<channel>
						<title>Different Blog</title>
						<item>
							<title>Different Post</title>
							<link>https://example.com/different-post</link>
							<description>Different test post</description>
							<enclosure url="https://example.com/image.jpg" type="image/jpeg"/>
						</item>
					</channel>
				</rss>`,
			statusCode:    200,
			expectedPosts: 1,
			expectedError: false,
			expectedTitle: "Different Post",
		},
		{
			name: "Malformed URL in RSS",
			xmlContent: `
				<rss>
					<channel>
						<title>Malformed Blog</title>
						<item>
							<title>Malformed Post</title>
							<link>invalid-url</link>
							<description>Malformed URL test</description>
						</item>
					</channel>
				</rss>`,
			statusCode:    200,
			expectedPosts: 1,
			expectedError: false,
			expectedTitle: "Malformed Post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockHTTPServer(tt.xmlContent, tt.statusCode)
			defer server.Close()

			posts, err := CheckRSSFeed(server.URL)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPosts, len(posts))

			if tt.expectedPosts > 0 && tt.expectedTitle != "" {
				assert.Equal(t, tt.expectedTitle, posts[0].Title)
			}
		})
	}
}

// Test hash content function
func TestHashContent(t *testing.T) {
	content := "This is a test post"
	actualHash := HashContent(content)

	expectedHash := [32]byte{171, 214, 38, 231, 215, 166, 144, 206, 157, 133, 112, 100, 123, 136, 149, 247, 102, 45, 79, 114, 7, 254, 136, 203, 103, 200, 223, 156, 18, 75, 167, 165}

	assert.Equal(t, expectedHash[:], actualHash[:])
}

// Helper function to mock an HTTP server
func mockHTTPServer(response string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(response))

	}))
}
