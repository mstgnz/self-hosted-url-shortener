package service

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/mstgnz/self-hosted-url-shortener/database"
	"github.com/mstgnz/self-hosted-url-shortener/model"
	"github.com/skip2/go-qrcode"
)

// URLService handles the business logic for URL shortening
type URLService struct {
	db *database.Database
}

// New creates a new URL service
func New(db *database.Database) *URLService {
	return &URLService{db: db}
}

// ShortenURL creates a shortened URL
func (s *URLService) ShortenURL(longURL, customCode string) (*model.URL, error) {
	// Validate the URL
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}

	var shortCode string
	if customCode != "" {
		// Check if the custom code is already in use
		existingURL, err := s.db.GetURLByShortCode(customCode)
		if err != nil {
			return nil, fmt.Errorf("error checking custom code: %w", err)
		}
		if existingURL != nil {
			return nil, fmt.Errorf("custom code '%s' is already in use", customCode)
		}
		shortCode = customCode
	} else {
		// Generate a random short code
		var err error
		shortCode, err = generateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		// Ensure the generated code is unique
		for {
			existingURL, err := s.db.GetURLByShortCode(shortCode)
			if err != nil {
				return nil, fmt.Errorf("error checking short code: %w", err)
			}
			if existingURL == nil {
				break
			}
			shortCode, err = generateShortCode(6)
			if err != nil {
				return nil, fmt.Errorf("failed to generate short code: %w", err)
			}
		}
	}

	// Create and save the URL
	url := model.NewURL(shortCode, longURL)
	if err := s.db.SaveURL(url); err != nil {
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	return url, nil
}

// GetURL retrieves a URL by its short code
func (s *URLService) GetURL(shortCode string) (*model.URL, error) {
	url, err := s.db.GetURLByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	return url, nil
}

// RecordClick records a click on a URL
func (s *URLService) RecordClick(shortCode string) error {
	return s.db.IncrementClicks(shortCode)
}

// ListURLs retrieves all URLs
func (s *URLService) ListURLs() ([]*model.URL, error) {
	return s.db.ListURLs()
}

// DeleteURL deletes a URL by its short code
func (s *URLService) DeleteURL(shortCode string) error {
	return s.db.DeleteURL(shortCode)
}

// GenerateQRCode generates a QR code for a URL
func (s *URLService) GenerateQRCode(shortURL string) ([]byte, error) {
	qr, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return qr, nil
}

// generateShortCode generates a random short code of the specified length
func generateShortCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i, b := range randomBytes {
		randomBytes[i] = charset[b%byte(len(charset))]
	}

	return string(randomBytes), nil
}
