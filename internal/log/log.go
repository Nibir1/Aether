// internal/log/log.go
//
// Package log provides a minimal logging abstraction for Aether.
// It wraps the standard library logger and exposes four severity levels.
//
// The purpose of this abstraction is to avoid forcing a specific logging
// framework on Aether users, while still enabling internal packages to
// emit useful diagnostics.
package log

import (
	stdlog "log"
	"os"
)

// Logger is the interface that Aether uses for logging.
//
// It is intentionally small so that it can be easily adapted to other
// logging frameworks if needed.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// Level represents the verbosity level of the logger.
type Level int

const (
	// LevelDebug enables all log messages.
	LevelDebug Level = iota
	// LevelInfo emits informational, warning and error messages.
	LevelInfo
	// LevelWarn emits only warnings and errors.
	LevelWarn
	// LevelError emits only errors.
	LevelError
)

// stdLogger is a simple implementation of Logger using the standard log package.
type stdLogger struct {
	level Level
	l     *stdlog.Logger
}

// New creates a new Logger instance.
//
// If debug is true, the logger will emit messages at LevelDebug;
// otherwise it uses LevelInfo as a reasonable default.
func New(debug bool) Logger {
	level := LevelInfo
	if debug {
		level = LevelDebug
	}

	return &stdLogger{
		level: level,
		l:     stdlog.New(os.Stderr, "[Aether] ", stdlog.LstdFlags|stdlog.Lmsgprefix),
	}
}

func (s *stdLogger) Debugf(format string, args ...any) {
	if s.level <= LevelDebug {
		s.l.Printf("DEBUG: "+format, args...)
	}
}

func (s *stdLogger) Infof(format string, args ...any) {
	if s.level <= LevelInfo {
		s.l.Printf("INFO: "+format, args...)
	}
}

func (s *stdLogger) Warnf(format string, args ...any) {
	if s.level <= LevelWarn {
		s.l.Printf("WARN: "+format, args...)
	}
}

func (s *stdLogger) Errorf(format string, args ...any) {
	if s.level <= LevelError {
		s.l.Printf("ERROR: "+format, args...)
	}
}
