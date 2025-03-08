package database

import "github.com/mstgnz/self-hosted-url-shortener/model"

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	// SaveURL saves a URL to the database
	SaveURL(url *model.URL) error

	// GetURLByShortCode retrieves a URL by its short code
	GetURLByShortCode(shortCode string) (*model.URL, error)

	// IncrementClicks increments the click count for a URL
	IncrementClicks(shortCode string) error

	// ListURLs returns all URLs in the database
	ListURLs() ([]*model.URL, error)

	// DeleteURL deletes a URL from the database
	DeleteURL(shortCode string) error

	// Close closes the database connection
	Close() error
}
