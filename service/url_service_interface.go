package service

import "github.com/mstgnz/self-hosted-url-shortener/model"

// URLServiceInterface defines the interface for URL service operations
type URLServiceInterface interface {
	// ShortenURL creates a shortened URL
	ShortenURL(longURL, customCode string) (*model.URL, error)

	// GetURL retrieves a URL by its short code
	GetURL(shortCode string) (*model.URL, error)

	// RecordClick records a click for a URL
	RecordClick(shortCode string) error

	// ListURLs returns all URLs
	ListURLs() ([]*model.URL, error)

	// DeleteURL deletes a URL
	DeleteURL(shortCode string) error

	// GenerateQRCode generates a QR code for a URL
	GenerateQRCode(shortURL string) ([]byte, error)
}
