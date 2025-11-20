// internal/normalize/doc.go
//
// Package normalize implements Aether's internal normalization pipeline.
//
// Normalization is the process of taking heterogeneous, source-specific
// structures (search results, extracted articles, feeds, entities, API
// responses) and converting them into a single, LLM-ready document model
// (internal/model.Document).
//
// High-level goals:
//
//   - Present a stable, source-agnostic schema to the public aether package.
//   - Encode rich structure (sections, metadata, provenance) in a form that
//     is easy for large language models to consume.
//   - Support both JSON and TOON serialization from the same canonical model.
//   - Allow multiple input types (SearchResult, Article, Feed, Entity) to be
//     merged into one normalized document when appropriate.
//   - Provide clear extension points for future plugins and transforms.
//
// Directory layout:
//
//	normalize.go
//	    Orchestrates the main normalization pipeline. Entry points here
//	    accept high-level types such as aether.SearchResult and delegate
//	    to schema-specific normalizers.
//
//	merge.go
//	    Contains logic for merging multiple normalized views (e.g. a primary
//	    search document plus article details plus feed items) into a single
//	    internal/model.Document instance.
//
//	schema_search.go
//	    Normalizes aether.SearchResult and its SearchDocument into the
//	    canonical document model. This is the main path used when callers
//	    ask Aether to normalize search results.
//
//	schema_article.go
//	    Normalizes Article structures from the readability/extractText
//	    subsystem into one or more body sections within a Document.
//
//	schema_feed.go
//	    Normalizes Feed and feed items (RSS/Atom) into sections with
//	    a "feed_item" role, preserving links, timestamps, and titles.
//
//	schema_entity.go
//	    Reserved for entity- and API-centric normalization (e.g. OpenAPI
//	    responses, Wikidata, or future structured entities). This allows
//	    Aether to represent non-article data in a consistent format.
//
//	util.go
//	    Shared helper functions such as whitespace normalization, safe
//	    metadata mapping, excerpt generation, and role constants.
//
// This package is intentionally internal. The public aether package exposes
// normalization via methods such as:
//
//	(*aether.Client).NormalizeSearchResult
//	(*aether.Client).MarshalSearchResultJSON
//	(*aether.Client).MarshalSearchResultTOON
//
// Those methods delegate into this package, which allows Aether's internal
// normalization logic to evolve without breaking the public API.
package normalize
