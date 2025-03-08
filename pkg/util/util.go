package util

import (
	"net/url"
	"strings"
	"time"
)

// IsValidURL checks if a URL is valid
func IsValidURL(rawURL string) bool {
	// Add scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Check if URL has a host
	return u.Host != ""
}

// FormatTime formats a time.Time value for display
func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CurrentYear returns the current year as a string
func CurrentYear() string {
	return time.Now().Format("2006")
}
