package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mstgnz/self-hosted-url-shortener/model"
)

// Database represents the SQLite database connection
type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection parameters
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Create the database instance
	database := &Database{db: db}

	// Initialize the database schema
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initSchema initializes the database schema
func (d *Database) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT UNIQUE NOT NULL,
		long_url TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		clicks INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`

	_, err := d.db.Exec(query)
	return err
}

// SaveURL saves a URL to the database
func (d *Database) SaveURL(url *model.URL) error {
	query := `
	INSERT INTO urls (short_code, long_url, created_at, clicks)
	VALUES (?, ?, ?, ?)
	`

	result, err := d.db.Exec(query, url.ShortCode, url.LongURL, url.CreatedAt, url.Clicks)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	url.ID = id
	return nil
}

// GetURLByShortCode retrieves a URL by its short code
func (d *Database) GetURLByShortCode(shortCode string) (*model.URL, error) {
	query := `
	SELECT id, short_code, long_url, created_at, clicks
	FROM urls
	WHERE short_code = ?
	`

	var url model.URL
	err := d.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.LongURL,
		&url.CreatedAt,
		&url.Clicks,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

// IncrementClicks increments the click count for a URL
func (d *Database) IncrementClicks(shortCode string) error {
	query := `
	UPDATE urls
	SET clicks = clicks + 1
	WHERE short_code = ?
	`

	_, err := d.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	return nil
}

// ListURLs retrieves all URLs from the database
func (d *Database) ListURLs() ([]*model.URL, error) {
	query := `
	SELECT id, short_code, long_url, created_at, clicks
	FROM urls
	ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		err := rows.Scan(
			&url.ID,
			&url.ShortCode,
			&url.LongURL,
			&url.CreatedAt,
			&url.Clicks,
		)
		if err != nil {
			log.Printf("Error scanning URL row: %v", err)
			continue
		}
		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating URL rows: %w", err)
	}

	return urls, nil
}

// DeleteURL deletes a URL by its short code
func (d *Database) DeleteURL(shortCode string) error {
	query := `
	DELETE FROM urls
	WHERE short_code = ?
	`

	_, err := d.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	return nil
}
