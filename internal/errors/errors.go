// internal/errors/errors.go
//
// Package errors defines reusable error types for Aether.
// Using structured errors allows callers to inspect and react to
// specific failure modes, such as configuration issues, HTTP failures,
// or robots.txt violations.
package errors

import "fmt"

// Kind represents a high-level category of error.
//
// Grouping errors by Kind makes it easier for callers to implement
// policies such as "retry on HTTP errors, but fail fast on robots.txt".
type Kind string

const (
	// KindUnknown represents an unspecified error category.
	KindUnknown Kind = "unknown"

	// KindConfig indicates a configuration-related error.
	KindConfig Kind = "config"

	// KindHTTP indicates an HTTP-related error.
	KindHTTP Kind = "http"

	// KindRobots indicates a robots.txt-related error.
	KindRobots Kind = "robots"

	// KindParsing indicates an error while parsing HTML, RSS, etc.
	KindParsing Kind = "parsing"
)

// Error is Aether's structured error type.
//
// It wraps a human-readable message and a Kind identifier so that callers
// can distinguish between different failure classes programmatically.
type Error struct {
	Kind Kind   // high-level category of the error
	Msg  string // descriptive message
	Err  error  // underlying error, if any
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Msg)
}

// Unwrap returns the underlying error, enabling errors.Is/As usage.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error with the provided kind and message.
//
// The underlying error may be nil if there is no nested error.
func New(kind Kind, msg string, underlying error) *Error {
	return &Error{
		Kind: kind,
		Msg:  msg,
		Err:  underlying,
	}
}
