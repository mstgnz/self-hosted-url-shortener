package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mstgnz/self-hosted-url-shortener/model"
	"github.com/mstgnz/self-hosted-url-shortener/service"
)

// MockURLService is a mock implementation of the URL service for testing
type MockURLService struct {
	urls map[string]*model.URL
	id   int64
}

// NewMockURLService creates a new mock URL service
func NewMockURLService() *MockURLService {
	return &MockURLService{
		urls: make(map[string]*model.URL),
		id:   1,
	}
}

// ShortenURL creates a shortened URL
func (m *MockURLService) ShortenURL(longURL, customCode string) (*model.URL, error) {
	shortCode := customCode
	if shortCode == "" {
		shortCode = "generated"
	}

	// Check if the custom code is already in use
	if _, exists := m.urls[shortCode]; exists && customCode != "" {
		return nil, fmt.Errorf("custom code '%s' is already in use", customCode)
	}

	url := &model.URL{
		ID:        m.id,
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
		Clicks:    0,
	}
	m.id++
	m.urls[shortCode] = url
	return url, nil
}

// GetURL retrieves a URL by its short code
func (m *MockURLService) GetURL(shortCode string) (*model.URL, error) {
	url, exists := m.urls[shortCode]
	if !exists {
		return nil, nil
	}
	return url, nil
}

// RecordClick records a click for a URL
func (m *MockURLService) RecordClick(shortCode string) error {
	url, exists := m.urls[shortCode]
	if !exists {
		return fmt.Errorf("URL with code '%s' not found", shortCode)
	}
	url.Clicks++
	return nil
}

// ListURLs returns all URLs
func (m *MockURLService) ListURLs() ([]*model.URL, error) {
	urls := make([]*model.URL, 0, len(m.urls))
	for _, url := range m.urls {
		urls = append(urls, url)
	}
	return urls, nil
}

// DeleteURL deletes a URL
func (m *MockURLService) DeleteURL(shortCode string) error {
	delete(m.urls, shortCode)
	return nil
}

// GenerateQRCode generates a QR code for a URL
func (m *MockURLService) GenerateQRCode(shortURL string) ([]byte, error) {
	return []byte("mock-qr-code"), nil
}

// Ensure MockURLService implements service.URLService interface
var _ service.URLServiceInterface = (*MockURLService)(nil)

func setupTestHandler(t *testing.T) (*HTTPHandler, *MockURLService) {
	mockService := NewMockURLService()

	// Create a mock template
	tmpl, err := template.New("base").Parse(`{{ define "content" }}{{ end }}{{ define "list" }}{{ end }}{{ define "result" }}{{ end }}`)
	if err != nil {
		t.Fatalf("Failed to create mock template: %v", err)
	}

	// Create handler directly without loading templates from disk
	handler := &HTTPHandler{
		urlService: mockService,
		baseURL:    "http://localhost:8080",
		templates:  tmpl,
	}

	return handler, mockService
}

func TestIndexHandler(t *testing.T) {
	t.Skip("Skipping test that requires template rendering")
}

func TestShortenURLHandler(t *testing.T) {
	t.Skip("Skipping test that requires template rendering")
}

func TestRedirectHandler(t *testing.T) {
	handler, service := setupTestHandler(t)

	// Create a URL
	url, err := service.ShortenURL("https://example.com", "test")
	if err != nil {
		t.Fatalf("Failed to shorten URL: %v", err)
	}

	// Create a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Set up the chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call the handler
	handler.redirectHandler(w, req)

	// Check the response
	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
	}
	if resp.Header.Get("Location") != url.LongURL {
		t.Errorf("Expected redirect to '%s', got '%s'", url.LongURL, resp.Header.Get("Location"))
	}
}

func TestAPIHandlers(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Test API shorten URL
	t.Run("APIShortenURL", func(t *testing.T) {
		// Create a JSON request
		reqBody := `{"url": "https://example.com", "custom_code": "api-test"}`
		req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Call the handler
		handler.apiShortenURLHandler(w, req)

		// Check the response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Parse the response
		var response map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if response["short_code"] != "api-test" {
			t.Errorf("Expected short code 'api-test', got '%v'", response["short_code"])
		}
		if response["long_url"] != "https://example.com" {
			t.Errorf("Expected long URL 'https://example.com', got '%v'", response["long_url"])
		}
	})

	// Test API list URLs
	t.Run("APIListURLs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/urls", nil)
		w := httptest.NewRecorder()

		// Call the handler
		handler.apiListURLsHandler(w, req)

		// Check the response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// Test API get URL
	t.Run("APIGetURL", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/url/api-test", nil)
		w := httptest.NewRecorder()

		// Set up the chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "api-test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the handler
		handler.apiGetURLHandler(w, req)

		// Check the response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Parse the response
		var response map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if response["short_code"] != "api-test" {
			t.Errorf("Expected short code 'api-test', got '%v'", response["short_code"])
		}
		if response["long_url"] != "https://example.com" {
			t.Errorf("Expected long URL 'https://example.com', got '%v'", response["long_url"])
		}
	})

	// Test API delete URL
	t.Run("APIDeleteURL", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/url/api-test", nil)
		w := httptest.NewRecorder()

		// Set up the chi context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("code", "api-test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call the handler
		handler.apiDeleteURLHandler(w, req)

		// Check the response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}
