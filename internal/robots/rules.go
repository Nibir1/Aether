// internal/robots/rules.go
//
// Package robots implements a minimal robots.txt parser and matcher.
// It supports User-agent, Allow and Disallow directives, and computes
// access rules based on the most specific matching path.
package robots

// Rule represents a single Allow or Disallow directive.
type Rule struct {
	Allow bool
	Path  string
}

// Group represents a User-agent group with associated rules.
type Group struct {
	Agents []string
	Rules  []Rule
}

// Robots is an in-memory representation of a robots.txt file.
type Robots struct {
	Groups []Group
}
