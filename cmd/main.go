package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mstgnz/self-hosted-url-shortener/database"
	"github.com/mstgnz/self-hosted-url-shortener/handler"
	"github.com/mstgnz/self-hosted-url-shortener/service"
)

var (
	// Command line flags
	port         = flag.Int("port", 8080, "HTTP server port")
	dbPath       = flag.String("db", "data.db", "SQLite database path")
	baseURL      = flag.String("base-url", "http://localhost:8080", "Base URL for shortened URLs")
	cliMode      = flag.Bool("cli", false, "Run in CLI mode")
	templatesDir = flag.String("templates", "templates", "Templates directory")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Create database connection
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create URL service
	urlService := service.New(db)

	// Check if running in CLI mode
	if *cliMode {
		// Create CLI handler
		cliHandler := handler.NewCLIHandler(urlService, *baseURL)
		rootCmd := cliHandler.SetupCommands()

		// Execute CLI command
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create HTTP handler
	httpHandler, err := handler.NewHTTPHandler(urlService, *baseURL, *templatesDir)
	if err != nil {
		log.Fatalf("Failed to create HTTP handler: %v", err)
	}

	// Create Chi router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	// Setup routes
	httpHandler.SetupRoutes(router)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting server on %s", addr)
	log.Printf("URL shortener available at %s", *baseURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
