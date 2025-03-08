package model

import (
	"time"
)

// URL represents a shortened URL in the database
type URL struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int64     `json:"clicks"`
}

// NewURL creates a new URL with default values
func NewURL(shortCode, longURL string) *URL {
	return &URL{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		Clicks:    0,
	}
}
