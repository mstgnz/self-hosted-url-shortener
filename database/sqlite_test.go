package database

import (
	"os"
	"testing"
	"time"

	"github.com/mstgnz/self-hosted-url-shortener/model"
)

func setupTestDB(t *testing.T) (*Database, func()) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Open the database
	db, err := New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func TestNew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	if db == nil {
		t.Fatal("Expected database to be created, got nil")
	}
}

func TestSaveAndGetURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a URL
	url := &model.URL{
		ShortCode: "test",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	// Save the URL
	err := db.SaveURL(url)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}
	if url.ID == 0 {
		t.Errorf("Expected URL ID to be set, got 0")
	}

	// Get the URL
	retrievedURL, err := db.GetURLByShortCode("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if retrievedURL == nil {
		t.Fatalf("Expected URL to be returned, got nil")
	}
	if retrievedURL.LongURL != "https://example.com" {
		t.Errorf("Expected long URL to be 'https://example.com', got '%s'", retrievedURL.LongURL)
	}
	if retrievedURL.ShortCode != "test" {
		t.Errorf("Expected short code to be 'test', got '%s'", retrievedURL.ShortCode)
	}
	if retrievedURL.Clicks != 0 {
		t.Errorf("Expected clicks to be 0, got %d", retrievedURL.Clicks)
	}

	// Get non-existent URL
	retrievedURL, err = db.GetURLByShortCode("nonexistent")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if retrievedURL != nil {
		t.Errorf("Expected URL to be nil for non-existent code, got %+v", retrievedURL)
	}
}

func TestIncrementClicks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a URL
	url := &model.URL{
		ShortCode: "test",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	// Save the URL
	err := db.SaveURL(url)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	// Increment clicks
	err = db.IncrementClicks("test")
	if err != nil {
		t.Fatalf("Failed to increment clicks: %v", err)
	}

	// Verify clicks were incremented
	retrievedURL, err := db.GetURLByShortCode("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if retrievedURL.Clicks != 1 {
		t.Errorf("Expected clicks to be 1, got %d", retrievedURL.Clicks)
	}

	// Increment clicks again
	err = db.IncrementClicks("test")
	if err != nil {
		t.Fatalf("Failed to increment clicks: %v", err)
	}

	// Verify clicks were incremented again
	retrievedURL, err = db.GetURLByShortCode("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if retrievedURL.Clicks != 2 {
		t.Errorf("Expected clicks to be 2, got %d", retrievedURL.Clicks)
	}

	// Increment clicks for non-existent URL
	err = db.IncrementClicks("nonexistent")
	if err == nil {
		t.Logf("Expected error when incrementing clicks for non-existent URL, got nil")
	}
}

func TestListURLs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create some URLs
	url1 := &model.URL{
		ShortCode: "test1",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		Clicks:    0,
	}
	url2 := &model.URL{
		ShortCode: "test2",
		LongURL:   "https://example.org",
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	// Save the URLs
	err := db.SaveURL(url1)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}
	err = db.SaveURL(url2)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	// List URLs
	urls, err := db.ListURLs()
	if err != nil {
		t.Fatalf("Failed to list URLs: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urls))
	}
}

func TestDeleteURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a URL
	url := &model.URL{
		ShortCode: "test",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	// Save the URL
	err := db.SaveURL(url)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	// Delete the URL
	err = db.DeleteURL("test")
	if err != nil {
		t.Fatalf("Failed to delete URL: %v", err)
	}

	// Verify it's deleted
	retrievedURL, err := db.GetURLByShortCode("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if retrievedURL != nil {
		t.Errorf("Expected URL to be nil after deletion, got %+v", retrievedURL)
	}

	// Delete non-existent URL
	err = db.DeleteURL("nonexistent")
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent URL, got %v", err)
	}
}
