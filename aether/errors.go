// aether/errors.go
//
// This file provides public accessors for Aether's structured error
// kinds. Internally, we use an internal/errors package to represent
// specific failure categories such as configuration issues or robots.txt
// violations. This wrapper re-exports those concepts in a stable way.
package aether

import internal "github.com/Nibir1/Aether/internal/errors"

// ErrorKind is a high-level category of Aether error.
//
// It is intentionally string-based so it can be logged, compared and
// inspected easily without depending on internal implementation details.
type ErrorKind = internal.Kind

// Public error kind constants that mirror the internal error kinds.
// These allow callers to distinguish between failure modes such as
// configuration errors, HTTP errors or robots.txt violations.
const (
	ErrorKindUnknown ErrorKind = internal.KindUnknown
	ErrorKindConfig  ErrorKind = internal.KindConfig
	ErrorKindHTTP    ErrorKind = internal.KindHTTP
	ErrorKindRobots  ErrorKind = internal.KindRobots
	ErrorKindParsing ErrorKind = internal.KindParsing
)

// Error is Aether's structured error type, re-exported for public use.
//
// Callers may use errors.Is / errors.As with this type, or inspect the
// Kind field directly to react to specific failure categories.
type Error = internal.Error
