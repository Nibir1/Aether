// plugins/interface.go
//
// Package plugins defines the public plugin interfaces for Aether.
//
// Plugins are an extension mechanism that allows applications to:
//   - Provide additional legal/public data sources (SourcePlugin)
//   - Enrich or transform normalized documents (TransformPlugin)
//   - Render documents into alternative formats (DisplayPlugin)
//
// IMPORTANT: Plugins must respect Aether's legal and ethical model.
// They MUST NOT:
//   - Bypass robots.txt
//   - Circumvent authentication or paywalled content
//   - Evade captchas or rate limits
//
// Source plugins should restrict themselves to:
//   - Public, openly accessible APIs
//   - Legal open data sources
//   - Aether's own high-level APIs (Search, OpenAPI wrappers, etc.)
// where possible rather than making raw HTTP calls that might violate
// the target site's policies.
//
// Aether core will provide registration functions and glue code in the
// aether package; this package only defines the interfaces and data
// structures that plugin authors implement.

package plugins

import "context"

//
// ────────────────────────────────────────────────
//              PLUGIN DOCUMENT MODEL
// ────────────────────────────────────────────────
//

// DocumentKind represents the broad category of a plugin-produced document.
//
// This is intentionally similar to Aether's internal normalized document
// kinds, but defined separately here to avoid import cycles and to keep
// the plugin API stable even if internals change.
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

// SectionRole is a free-form role label for a document section.
// Common examples: "body", "summary", "feed_item", "metadata".
type SectionRole string

// Section is a logical chunk of content within a plugin Document,
// such as an article body, a feed item, or a metadata block.
type Section struct {
	Role  SectionRole       `json:"role,omitempty"`
	Title string            `json:"title,omitempty"`
	Text  string            `json:"text,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
}

// Document is the main data structure plugins work with.
//
// It is intentionally similar (but not identical) to Aether's internal
// normalized document model. Aether will convert between this type and
// its internal representation during registration / execution.
//
// Plugin authors should focus on filling in as much useful, LLM-friendly
// content as possible (Title, Excerpt, Content, Sections, Metadata).
type Document struct {
	// Source is a human-readable identifier for the origin of this
	// document, e.g. "plugin:hackernews", "plugin:my_api", etc.
	Source string `json:"source,omitempty"`

	// URL is an optional canonical URL for the document, if any.
	URL string `json:"url,omitempty"`

	// Kind is the broad category of the document (article, feed, text, etc.).
	Kind DocumentKind `json:"kind,omitempty"`

	// Title is the main title or name of the document.
	Title string `json:"title,omitempty"`

	// Excerpt is a short summary or teaser.
	Excerpt string `json:"excerpt,omitempty"`

	// Content is the primary body text, if there is a single dominant
	// body for the document.
	Content string `json:"content,omitempty"`

	// Metadata is an arbitrary, flat key/value map for additional data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Sections is an optional list of structured sections, such as
	// article body blocks, feed entries, or metadata sections.
	Sections []Section `json:"sections,omitempty"`
}

//
// ────────────────────────────────────────────────
//                SOURCE PLUGINS
// ────────────────────────────────────────────────
//

// SourcePlugin is responsible for providing new documents from legal,
// public data sources. Examples:
//
//   - A Hacker News source plugin
//   - A public government dataset plugin
//   - A custom internal API plugin (when used in a private deployment)
//
// SourcePlugin implementations should be stateless or internally safe
// for concurrent use.
type SourcePlugin interface {
	// Name returns a short, stable identifier for the plugin.
	// Example: "hackernews", "my_org_api", "eu_press".
	Name() string

	// Description returns a human-readable explanation of what the
	// plugin does, which may be surfaced in logs or introspection.
	Description() string

	// Capabilities declares what kinds of queries or intents this
	// plugin is good at handling. Examples:
	//
	//   []string{"news", "tech", "hn"}
	//   []string{"weather"}
	//
	// Aether’s SmartQuery router can use these tags to decide when
	// to invoke the plugin.
	Capabilities() []string

	// Fetch executes a plugin-specific retrieval based on the incoming
	// query. The query may be a free-form question, a keyword, or a
	// structured string depending on how the plugin is used.
	//
	// The returned Document should be as normalized and clean as the
	// plugin can reasonably make it.
	Fetch(ctx context.Context, query string) (*Document, error)
}

//
// ────────────────────────────────────────────────
//              TRANSFORM PLUGINS
// ────────────────────────────────────────────────
//

// TransformPlugin receives an existing Document and returns a new (or
// modified) Document. Examples:
//
//   - Post-hoc summarization
//   - Keyword extraction
//   - Entity extraction
//   - Additional metadata enrichment
//
// Transform plugins must be pure functions in spirit: given the same
// input Document, they should produce the same output Document,
// barring non-deterministic external services.
type TransformPlugin interface {
	// Name returns a stable identifier for this transform plugin.
	Name() string

	// Description describes what transformation this plugin performs.
	Description() string

	// Apply transforms the input Document and returns the resulting
	// Document. Implementations may modify the input in place or
	// allocate a new Document, but must document their behavior.
	Apply(ctx context.Context, doc *Document) (*Document, error)
}

//
// ────────────────────────────────────────────────
//              DISPLAY PLUGINS (FORMATTERS)
// ────────────────────────────────────────────────
//

// DisplayPlugin renders a Document into a non-Markdown format.
//
// Examples:
//   - ANSI-colored CLI view
//   - HTML view
//   - PDF output (via a legal PDF library)
//   - TOON-derived visual representations
//
// Stage 13 defines the interface, but Aether will connect this into
// the Display subsystem and CLI/TUI layers in later stages.
type DisplayPlugin interface {
	// Name returns a stable identifier for the display plugin.
	Name() string

	// Description explains what output this plugin produces.
	Description() string

	// Format returns a short format tag such as:
	//   "ansi", "html", "pdf", "text", "custom"
	Format() string

	// Render converts the given Document into the plugin's output
	// format. The returned byte slice may contain text, HTML, binary
	// document data, etc., depending on Format().
	Render(ctx context.Context, doc *Document) ([]byte, error)
}
