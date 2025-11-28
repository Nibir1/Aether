// aether/errors.go
//
// Public accessors for Aether's structured error system.
//
// Internally, Aether uses internal/errors to classify failure modes
// (configuration issues, HTTP failures, robots.txt violations, parsing
// errors, etc.). This file re-exports those types and constants in a
// stable public API so that callers can:
//
//   • inspect error kinds via Error.Kind
//   • use errors.Is / errors.As with *aether.Error
//   • match specific failure categories without importing internal packages
//
// The public API intentionally mirrors internal error kinds, but keeps
// the freedom to expand internally without breaking user code.

package aether

import (
	internal "github.com/Nibir1/Aether/internal/errors"
)

//
// ───────────────────────────────────────────────────────────────
//                          ERROR KIND
// ───────────────────────────────────────────────────────────────
//
// ErrorKind classifies structured failure categories such as:
//
//   • config errors
//   • HTTP & transport errors
//   • robots.txt violations
//   • parsing errors
//
// This is re-exported from internal/errors.Kind.
// Using a type alias guarantees binary and semantic compatibility.

type ErrorKind = internal.Kind

// Publicly visible error kind constants mirroring internal.
//
// These allow callers to write:
//
//	if err, ok := err.(*aether.Error); ok && err.Kind == aether.ErrorKindHTTP { … }
//
// without importing internal/errors.
const (
	ErrorKindUnknown ErrorKind = internal.KindUnknown
	ErrorKindConfig  ErrorKind = internal.KindConfig
	ErrorKindHTTP    ErrorKind = internal.KindHTTP
	ErrorKindRobots  ErrorKind = internal.KindRobots
	ErrorKindParsing ErrorKind = internal.KindParsing
)

// ───────────────────────────────────────────────────────────────
//
//	STRUCTURED ERROR
//
// ───────────────────────────────────────────────────────────────
//
// Error is the structured error type returned by many Aether functions.
//
// It contains:
//   - Kind        (error category)
//   - Op          (optional operation name)
//   - URL         (optional relevant URL)
//   - Err         (wrapped underlying error)
//
// Users may pattern-match with:
//
//	if ae, ok := err.(*aether.Error); ok {
//	    switch ae.Kind {
//	    case aether.ErrorKindRobots: …
type Error = internal.Error
