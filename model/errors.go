package model

import "fmt"

// ErrCustomCodeAlreadyExists is returned when a custom code is already in use
type ErrCustomCodeAlreadyExists struct {
	Code string
}

// Error returns the error message
func (e *ErrCustomCodeAlreadyExists) Error() string {
	return fmt.Sprintf("custom code '%s' is already in use", e.Code)
}

// ErrURLNotFound is returned when a URL is not found
type ErrURLNotFound struct {
	Code string
}

// Error returns the error message
func (e *ErrURLNotFound) Error() string {
	return fmt.Sprintf("URL with code '%s' not found", e.Code)
}

// ErrInvalidURL is returned when a URL is invalid
type ErrInvalidURL struct {
	URL string
}

// Error returns the error message
func (e *ErrInvalidURL) Error() string {
	return fmt.Sprintf("invalid URL: %s", e.URL)
}

// ErrDatabaseError is returned when a database error occurs
type ErrDatabaseError struct {
	Err error
}

// Error returns the error message
func (e *ErrDatabaseError) Error() string {
	return fmt.Sprintf("database error: %v", e.Err)
}
