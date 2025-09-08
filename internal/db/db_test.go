package db

import (
	"os"
	"testing"
)

// Test initializing the DB
func TestInitDB(t *testing.T) {
	InitDB()
	defer CloseDB()

	_, err := DB.Exec("SELECT 1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test storing a new post
func TestStoreTootedPost_NewPost(t *testing.T) {
	InitDB()
	defer CloseDB()

	err := StoreTootedPost("https://example.com/test-post", "Test post content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test checking if a post has changed (new post case)
func TestHasPostChanged_NewPost(t *testing.T) {
	InitDB()
	defer CloseDB()

	exists, updated, err := HasPostChanged("https://example.com/test-post-2", "Test post 2 content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if exists {
		t.Errorf("Expected post to be new")
	}

	if updated {
		t.Errorf("Expected post to be new, not updated")
	}
}

// Test checking if a post has changed (updated post case)
func TestHasPostChanged_UpdatedPost(t *testing.T) {
	InitDB()
	defer CloseDB()

	// Insert a post
	err := StoreTootedPost("https://example.com/test-post", "Original content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Now check with updated content
	exists, updated, err := HasPostChanged("https://example.com/test-post", "Updated content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !exists {
		t.Errorf("Expected post to exist")
	}

	if !updated {
		t.Errorf("Expected post to be updated")
	}
}

// Test case where post exists but is not updated
func TestHasPostChanged_UnchangedPost(t *testing.T) {
	InitDB()
	defer CloseDB()

	// Insert a post
	err := StoreTootedPost("https://example.com/test-post", "Test post content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check again with the same content
	exists, updated, err := HasPostChanged("https://example.com/test-post", "Test post content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !exists {
		t.Errorf("Expected post to exist")
	}

	if updated {
		t.Errorf("Expected post to be unchanged")
	}
}

// Clean up test database
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup test database file
	os.Remove("./tooted_posts.db")

	os.Exit(code)
}
