// Package apperrors holds sentinel errors shared across feature packages
// (blog, comment, ...) so the HTTP layer can map any of them to a status
// code via a single errors.Is check, without importing each feature
// package just for its error variable.
package apperrors

import "errors"

var (
	// ErrNotFound indicates the requested resource (a blog, a comment, ...)
	// does not exist. Maps to HTTP 404.
	ErrNotFound = errors.New("not found")

	// ErrValidation indicates the caller's input failed validation. Wrap it
	// with fmt.Errorf("%w: <detail>", ErrValidation) so the detail can be
	// surfaced to the caller while errors.Is(err, ErrValidation) still
	// matches. Maps to HTTP 400.
	ErrValidation = errors.New("validation failed")
)
