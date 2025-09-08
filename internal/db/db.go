// Package db provides database operations for storing and checking tooted posts using SQLite.
// It manages the database connection, table creation, and CRUD operations for post tracking.
package db

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
	"github.com/toozej/rss2socials/internal/rss"
)

var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB() {
	// InitDB opens the SQLite database file and creates the tooted_posts table if it doesn't exist.
	// It sets up the schema for storing post links, content hashes, and timestamps.
	var err error
	DB, err = sql.Open("sqlite3", "./tooted_posts.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create table if not exists
	query := `CREATE TABLE IF NOT EXISTS tooted_posts (
		link TEXT PRIMARY KEY,
		content_hash TEXT,
		timestamp TEXT
	)`
	_, err = DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

// CloseDB closes the SQLite database connection
func CloseDB() {
	// CloseDB closes the active SQLite database connection.
	err := DB.Close()
	if err != nil {
		log.Error("Error closing SQLite database connection: ", err)
	}
}

// StoreTootedPost stores the link, content hash, and timestamp in the database
func StoreTootedPost(link string, content string) error {
	// StoreTootedPost inserts or replaces a post record in the database with its link, content hash, and current timestamp.
	// It uses the RSS hash function to compute the content hash.
	query := `INSERT OR REPLACE INTO tooted_posts(link, content_hash, timestamp) VALUES (?, ?, ?)`
	contentHash := rss.HashContent(content)
	_, err := DB.Exec(query, link, fmt.Sprintf("%x", contentHash), time.Now().Format(time.RFC3339))
	return err
}

// HasPostChanged checks if the post content has changed or if it is new
func HasPostChanged(link string, content string) (exists bool, updated bool, err error) {
	// HasPostChanged checks if a post with the given link exists in the database and if its content has changed.
	// Returns exists (true if post found), updated (true if content differs), and any error.
	query := `SELECT content_hash FROM tooted_posts WHERE link = ?`
	row := DB.QueryRow(query, link)

	var storedHash string
	err = row.Scan(&storedHash)
	if err == sql.ErrNoRows {
		// Post is new
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Check if the content hash has changed
	newHash := fmt.Sprintf("%x", rss.HashContent(content))
	if storedHash != newHash {
		// Post has been updated
		return true, true, nil
	}

	// Post already exists and is unchanged
	return true, false, nil
}
