package service

import (
	"os"
	"testing"

	"github.com/mstgnz/self-hosted-url-shortener/database"
	"github.com/mstgnz/self-hosted-url-shortener/model"
)

// MockDatabase implements the database interface for testing
type MockDatabase struct {
	urls map[string]*model.URL
	id   int64
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		urls: make(map[string]*model.URL),
		id:   1,
	}
}

// SaveURL saves a URL to the mock database
func (m *MockDatabase) SaveURL(url *model.URL) error {
	url.ID = m.id
	m.id++
	m.urls[url.ShortCode] = url
	return nil
}

// GetURLByShortCode retrieves a URL by its short code
func (m *MockDatabase) GetURLByShortCode(shortCode string) (*model.URL, error) {
	url, exists := m.urls[shortCode]
	if !exists {
		return nil, nil
	}
	return url, nil
}

// IncrementClicks increments the click count for a URL
func (m *MockDatabase) IncrementClicks(shortCode string) error {
	url, exists := m.urls[shortCode]
	if !exists {
		return os.ErrNotExist
	}
	url.Clicks++
	return nil
}

// ListURLs returns all URLs in the mock database
func (m *MockDatabase) ListURLs() ([]*model.URL, error) {
	urls := make([]*model.URL, 0, len(m.urls))
	for _, url := range m.urls {
		urls = append(urls, url)
	}
	return urls, nil
}

// DeleteURL deletes a URL from the mock database
func (m *MockDatabase) DeleteURL(shortCode string) error {
	delete(m.urls, shortCode)
	return nil
}

// Close is a no-op for the mock database
func (m *MockDatabase) Close() error {
	return nil
}

// Ensure MockDatabase implements database.Database interface
var _ database.DatabaseInterface = (*MockDatabase)(nil)

func TestShortenURL(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Test with custom code
	url, err := service.ShortenURL("https://example.com", "custom")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}
	if url.LongURL != "https://example.com" {
		t.Errorf("Expected long URL to be 'https://example.com', got '%s'", url.LongURL)
	}
	if url.ShortCode != "custom" {
		t.Errorf("Expected short code to be 'custom', got '%s'", url.ShortCode)
	}
	if url.Clicks != 0 {
		t.Errorf("Expected clicks to be 0, got %d", url.Clicks)
	}

	// Test without custom code
	url, err = service.ShortenURL("https://example.org", "")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}
	if url.LongURL != "https://example.org" {
		t.Errorf("Expected long URL to be 'https://example.org', got '%s'", url.LongURL)
	}
	if url.ShortCode == "" {
		t.Errorf("Expected short code to be generated, got empty string")
	}
	if url.Clicks != 0 {
		t.Errorf("Expected clicks to be 0, got %d", url.Clicks)
	}

	// Test duplicate custom code
	_, err = service.ShortenURL("https://example.net", "custom")
	if err == nil {
		t.Errorf("Expected error for duplicate custom code, got nil")
	}
}

func TestGetURL(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Create a URL
	_, err := service.ShortenURL("https://example.com", "test")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// Get the URL
	url, err := service.GetURL("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if url == nil {
		t.Fatalf("Expected URL to be returned, got nil")
	}
	if url.LongURL != "https://example.com" {
		t.Errorf("Expected long URL to be 'https://example.com', got '%s'", url.LongURL)
	}
	if url.ShortCode != "test" {
		t.Errorf("Expected short code to be 'test', got '%s'", url.ShortCode)
	}

	// Get non-existent URL
	url, err = service.GetURL("nonexistent")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if url != nil {
		t.Errorf("Expected URL to be nil for non-existent code, got %+v", url)
	}
}

func TestListURLs(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Create some URLs
	_, err := service.ShortenURL("https://example.com", "test1")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}
	_, err = service.ShortenURL("https://example.org", "test2")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// List URLs
	urls, err := service.ListURLs()
	if err != nil {
		t.Fatalf("Failed to list URLs: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urls))
	}
}

func TestDeleteURL(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Create a URL
	_, err := service.ShortenURL("https://example.com", "test")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// Delete the URL
	err = service.DeleteURL("test")
	if err != nil {
		t.Fatalf("Failed to delete URL: %v", err)
	}

	// Verify it's deleted
	url, err := service.GetURL("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if url != nil {
		t.Errorf("Expected URL to be nil after deletion, got %+v", url)
	}
}

func TestRecordClick(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Create a URL
	_, err := service.ShortenURL("https://example.com", "test")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// Record a click
	err = service.RecordClick("test")
	if err != nil {
		t.Fatalf("Failed to record click: %v", err)
	}

	// Verify click was recorded
	url, err := service.GetURL("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if url.Clicks != 1 {
		t.Errorf("Expected clicks to be 1, got %d", url.Clicks)
	}

	// Record another click
	err = service.RecordClick("test")
	if err != nil {
		t.Fatalf("Failed to record click: %v", err)
	}

	// Verify click was recorded
	url, err = service.GetURL("test")
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}
	if url.Clicks != 2 {
		t.Errorf("Expected clicks to be 2, got %d", url.Clicks)
	}

	// Record click for non-existent URL
	err = service.RecordClick("nonexistent")
	if err == nil {
		t.Errorf("Expected error when recording click for non-existent URL, got nil")
	}
}

func TestGenerateQRCode(t *testing.T) {
	mockDB := NewMockDatabase()
	service := New(mockDB)

	// Generate QR code
	qrCode, err := service.GenerateQRCode("https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate QR code: %v", err)
	}
	if len(qrCode) == 0 {
		t.Errorf("Expected QR code to be generated, got empty byte slice")
	}
}
