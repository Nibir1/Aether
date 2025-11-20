// internal/model/document.go
//
// Package model defines Aether's canonical normalized document model.
// All high-level outputs (Search, Article extraction, RSS, OpenAPI,
// Entities, Plugins) are converted into this structure before final
// JSON or TOON serialization.
//
// The purpose of this package is to provide a single, stable schema
// that is easy for applications and LLMs to consume, while remaining
// flexible enough to represent:
//   - Articles (HTML → readable content)
//   - Feeds (RSS/Atom)
//   - JSON or text pages
//   - Binary or unsupported content
//   - Structured entities (Wikidata, OpenAPI GOV data, etc.)
//   - Plugin-defined structured objects

package model

//
// ──────────────────────────────────────────────────────────────────────────
//                               SECTION ROLES
// ──────────────────────────────────────────────────────────────────────────
//

// SectionRole describes the semantic purpose of a section.
type SectionRole string

const (
	// Traditional article sections
	SectionRoleBody    SectionRole = "body"
	SectionRoleSummary SectionRole = "summary"

	// RSS/Atom feed items
	SectionRoleFeedItem SectionRole = "feed_item"

	// Structured entities (Wikidata, gov APIs, plugin entities)
	SectionRoleEntity SectionRole = "entity"

	// Special metadata sections (fallback)
	SectionRoleMetadata SectionRole = "metadata"

	// Unknown / unspecified
	SectionRoleUnknown SectionRole = "unknown"
)

// Section represents a logical content unit within a document.
// Examples:
//   - article body section
//   - feed item
//   - structured entity (key/values)
//   - metadata notes
type Section struct {
	Role    SectionRole       `json:"role,omitempty"`
	Heading string            `json:"heading,omitempty"`
	Text    string            `json:"text,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

//
// ──────────────────────────────────────────────────────────────────────────
//                              DOCUMENT KINDS
// ──────────────────────────────────────────────────────────────────────────
//

// DocumentKind classifies the normalized document at a high level.
type DocumentKind string

const (
	// Default / unknown
	DocumentKindUnknown DocumentKind = "unknown"

	// Article-like content (readability extraction)
	DocumentKindArticle DocumentKind = "article"

	// HTML page (non-article or generic)
	DocumentKindHTML DocumentKind = "html_page"

	// RSS/Atom feeds
	DocumentKindFeed DocumentKind = "feed"

	// JSON APIs or JSON resources
	DocumentKindJSON DocumentKind = "json"

	// Plain text pages
	DocumentKindText DocumentKind = "text"

	// Binary or unsupported content
	DocumentKindBinary DocumentKind = "binary"

	// Structured entity (Wikidata, Gov APIs, Marketplace APIs, Plugins)
	DocumentKindEntity DocumentKind = "entity"
)

//
// ──────────────────────────────────────────────────────────────────────────
//                                DOCUMENT
// ──────────────────────────────────────────────────────────────────────────
//

// Document is Aether's canonical normalized representation.
//
// Every high-level result—SearchResult, Article, Feed, Entity, JSON API,
// plugin output—gets converted into this structure before final
// serialization to JSON or TOON.
//
// Applications consuming Aether should operate on this type rather than
// on source-specific types (SearchResult, Article, Feed, etc.).
type Document struct {
	// Basic identity + source reference
	SourceURL string       `json:"source_url,omitempty"`
	Kind      DocumentKind `json:"kind"`

	// Core content
	Title   string `json:"title,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
	Content string `json:"content,omitempty"`

	// Arbitrary key/value metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// Structured content
	// Examples:
	//   - article body paragraphs
	//   - feed items
	//   - structured entity fields
	//   - metadata blocks
	Sections []Section `json:"sections,omitempty"`
}
