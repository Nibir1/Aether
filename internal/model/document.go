// internal/model/document.go
//
// Package model defines Aether's internal normalized document model.
// This is the canonical, implementation-agnostic representation that
// all higher-level formats (JSON, TOON, etc.) are derived from.
//
// The goal is to have a single, stable structure that can represent
// the output of Search, ExtractArticle, RSS parsing, and OpenAPI
// integrations in a way that is LLM-friendly and easy to serialize.

package model

// SectionRole describes the semantic role of a section.
type SectionRole string

const (
	SectionRoleBody     SectionRole = "body"
	SectionRoleSummary  SectionRole = "summary"
	SectionRoleFeedItem SectionRole = "feed_item"
	SectionRoleMetadata SectionRole = "metadata"
	SectionRoleUnknown  SectionRole = "unknown"
)

// Section represents a logical chunk of content within a document.
// For example, an article body, a feed item summary, or a metadata note.
type Section struct {
	Role    SectionRole       `json:"role,omitempty"`
	Heading string            `json:"heading,omitempty"`
	Text    string            `json:"text,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// DocumentKind is the high-level kind of normalized document.
type DocumentKind string

const (
	DocumentKindUnknown DocumentKind = "unknown"
	DocumentKindArticle DocumentKind = "article"
	DocumentKindHTML    DocumentKind = "html_page"
	DocumentKindFeed    DocumentKind = "feed"
	DocumentKindJSON    DocumentKind = "json"
	DocumentKindText    DocumentKind = "text"
	DocumentKindBinary  DocumentKind = "binary"
)

// Document is Aether's canonical normalized document.
//
// Every SearchResult, article extraction, feed, or OpenAPI response
// that is intended for LLM consumption should be convertible into
// this structure.
type Document struct {
	// Basic identity
	SourceURL string       `json:"source_url,omitempty"`
	Kind      DocumentKind `json:"kind"`

	// Core content
	Title   string `json:"title,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
	Content string `json:"content,omitempty"`

	// Arbitrary metadata (flattened key/value map)
	Metadata map[string]string `json:"metadata,omitempty"`

	// Optional structured sections (article body, feed entries, etc.)
	Sections []Section `json:"sections,omitempty"`
}
