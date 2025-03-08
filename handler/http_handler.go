package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mstgnz/self-hosted-url-shortener/service"
)

// Define a custom key type to avoid collisions in context values
type contextKey string

// Define context keys
const currentYearKey contextKey = "currentYear"

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	urlService service.URLServiceInterface
	baseURL    string
	templates  *template.Template
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(urlService service.URLServiceInterface, baseURL string, templatesDir string) (*HTTPHandler, error) {
	// Load templates with base template first
	templates := template.New("")

	// Add helper functions
	templates = templates.Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},
	})

	// Parse all templates
	templates, err := templates.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &HTTPHandler{
		urlService: urlService,
		baseURL:    baseURL,
		templates:  templates,
	}, nil
}

// SetupRoutes sets up the HTTP routes
func (h *HTTPHandler) SetupRoutes(router chi.Router) {
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(h.currentYearMiddleware)

	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Handle favicon.ico to prevent 404 errors
	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	})

	// Web interface routes
	router.Get("/", h.indexHandler)
	router.Get("/urls", h.listURLsHandler)
	router.Post("/shorten", h.shortenURLHandler)
	router.Get("/qr/{code}", h.qrCodeHandler)
	router.Get("/delete/{code}", h.deleteURLHandler)

	// API routes
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", h.apiShortenURLHandler)
		r.Get("/urls", h.apiListURLsHandler)
		r.Get("/url/{code}", h.apiGetURLHandler)
		r.Delete("/url/{code}", h.apiDeleteURLHandler)
	})

	// Redirect route
	router.Get("/{code}", h.redirectHandler)
}

// currentYearMiddleware adds the current year to the request context
func (h *HTTPHandler) currentYearMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, currentYearKey, time.Now().Format("2006"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// indexHandler handles the index page
func (h *HTTPHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "base.html", map[string]any{
		"baseURL":     h.baseURL,
		"currentYear": r.Context().Value(currentYearKey),
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}

// listURLsHandler handles the URL listing page
func (h *HTTPHandler) listURLsHandler(w http.ResponseWriter, r *http.Request) {
	urls, err := h.urlService.ListURLs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list URLs: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.templates.ExecuteTemplate(w, "base.html", map[string]any{
		"urls":        urls,
		"baseURL":     h.baseURL,
		"currentYear": r.Context().Value(currentYearKey),
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}

// shortenURLHandler handles URL shortening requests
func (h *HTTPHandler) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid form data: %v", err), http.StatusBadRequest)
		return
	}

	longURL := r.PostForm.Get("url")
	customCode := r.PostForm.Get("custom_code")

	if longURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	url, err := h.urlService.ShortenURL(longURL, customCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error shortening URL: %v", err), http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)

	err = h.templates.ExecuteTemplate(w, "base.html", map[string]any{
		"url":         url,
		"shortURL":    shortURL,
		"baseURL":     h.baseURL,
		"currentYear": r.Context().Value(currentYearKey),
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}

// qrCodeHandler generates a QR code for a URL
func (h *HTTPHandler) qrCodeHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, code)

	qrCode, err := h.urlService.GenerateQRCode(shortURL)
	if err != nil {
		h.templates.ExecuteTemplate(w, "error.html", map[string]any{
			"error":       "Failed to generate QR code",
			"currentYear": r.Context().Value(currentYearKey),
		})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(qrCode)
}

// deleteURLHandler handles URL deletion requests
func (h *HTTPHandler) deleteURLHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	err := h.urlService.DeleteURL(code)
	if err != nil {
		h.templates.ExecuteTemplate(w, "error.html", map[string]any{
			"error":       "Failed to delete URL",
			"currentYear": r.Context().Value(currentYearKey),
		})
		return
	}

	http.Redirect(w, r, "/urls", http.StatusFound)
}

// redirectHandler redirects to the original URL
func (h *HTTPHandler) redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Special case for favicon.ico to prevent errors
	if code == "favicon.ico" {
		http.ServeFile(w, r, "./static/favicon.ico")
		return
	}

	url, err := h.urlService.GetURL(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve URL: %v", err), http.StatusInternalServerError)
		return
	}

	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// Record click asynchronously
	go h.urlService.RecordClick(code)

	http.Redirect(w, r, url.LongURL, http.StatusFound)
}

// API Handlers

// apiShortenURLHandler handles API URL shortening requests
func (h *HTTPHandler) apiShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL        string `json:"url"`
		CustomCode string `json:"custom_code"`
	}

	// Parse JSON request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	url, err := h.urlService.ShortenURL(request.URL, request.CustomCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)

	response := map[string]any{
		"id":         url.ID,
		"short_code": url.ShortCode,
		"long_url":   url.LongURL,
		"short_url":  shortURL,
		"created_at": url.CreatedAt.Format(time.RFC3339),
		"clicks":     url.Clicks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// apiListURLsHandler handles API URL listing requests
func (h *HTTPHandler) apiListURLsHandler(w http.ResponseWriter, r *http.Request) {
	urls, err := h.urlService.ListURLs()
	if err != nil {
		http.Error(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	var response []map[string]any
	for _, url := range urls {
		shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)
		response = append(response, map[string]any{
			"id":         url.ID,
			"short_code": url.ShortCode,
			"long_url":   url.LongURL,
			"short_url":  shortURL,
			"created_at": url.CreatedAt.Format(time.RFC3339),
			"clicks":     url.Clicks,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"urls": response})
}

// apiGetURLHandler handles API URL retrieval requests
func (h *HTTPHandler) apiGetURLHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	url, err := h.urlService.GetURL(code)
	if err != nil {
		http.Error(w, "Failed to retrieve URL", http.StatusInternalServerError)
		return
	}

	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)

	response := map[string]any{
		"id":         url.ID,
		"short_code": url.ShortCode,
		"long_url":   url.LongURL,
		"short_url":  shortURL,
		"created_at": url.CreatedAt.Format(time.RFC3339),
		"clicks":     url.Clicks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// apiDeleteURLHandler handles API URL deletion requests
func (h *HTTPHandler) apiDeleteURLHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	err := h.urlService.DeleteURL(code)
	if err != nil {
		http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "URL deleted successfully"})
}
