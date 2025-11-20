// Package plugins defines Aether’s public plugin architecture.
//
// Aether plugins allow developers to extend the Aether system without
// modifying its core. Plugins are optional, modular components that
// provide additional data sources, transformations, or renderers.
//
// Plugins MUST respect Aether’s legal and ethical principles:
//   - Robots.txt compliance
//   - No CAPTCHA bypassing
//   - No authentication circumvention
//   - No scraping of disallowed or paywalled content
//
// Aether exposes three primary plugin types:
//
//  1. SourcePlugin
//     ----------------
//     A SourcePlugin provides new legal/public data sources. It receives
//     a free-form query string and returns a plugins.Document representing
//     the fetched content. Source plugins integrate with Aether’s
//     SmartQuery routing engine and high-level Search pipeline.
//
//     Examples:
//     • Custom Hacker News retrieval
//     • Open government datasets
//     • Legally accessible public JSON APIs
//     • Local filesystem loaders
//
//  2. TransformPlugin
//     ----------------
//     A TransformPlugin receives a normalized plugins.Document and returns
//     a modified/enriched version. Transform plugins are ideal for:
//
//     • Metadata enrichment
//     • Summaries
//     • Entity extraction
//     • Keyword extraction
//
//     Transform plugins are applied after the primary Source pipeline
//     and before final display output.
//
//  3. DisplayPlugin
//     ----------------
//     A DisplayPlugin renders a plugins.Document into an alternative
//     output format. Markdown is built into Aether’s Display subsystem,
//     but DisplayPlugins allow:
//
//     • ANSI-styled CLI output
//     • HTML rendering
//     • TOON visualization
//     • PDF generation (via legal libraries)
//
//     Display plugins are connected to the Aether Client but operate
//     outside the Markdown core.
//
// # Plugin Registration
//
// Plugins are registered through methods exposed on the *aether.Client,
// for example:
//
//	cli := aether.NewClient(...)
//	cli.RegisterSourcePlugin(myPlugin)
//	cli.RegisterTransformPlugin(enricher)
//	cli.RegisterDisplayPlugin(htmlRenderer)
//
// Aether maintains a thread-safe plugin registry. Source plugins are
// invoked based on SmartQuery routing rules or explicit use. Transform
// plugins apply to normalized outputs. Display plugins expose new output
// formats.
//
// # Plugin Safety Model
//
// Plugins must never:
//   - Modify Aether’s HTTP fetcher
//   - Perform raw HTTP calls that ignore robots.txt
//   - Attempt to bypass protections or access restricted content
//
// Plugins MAY:
//   - Call public APIs that explicitly allow programmatic access
//   - Use Aether’s own high-level APIs (Search, OpenAPI, RSS)
//   - Operate entirely on normalized documents
//
// Aether enforces these rules by controlling fetch operations:
// plugins do not receive direct access to the internal HTTP client.
//
// # Plugin Document Model
//
// Plugins work with the plugins.Document structure. This structure is
// intentionally similar to internal/model.Document, but independent of
// internal implementation details. Aether converts between these two
// formats during plugin execution.
//
// The plugin Document format is intentionally LLM-friendly and structured.
//
// # Purpose and Philosophy
//
// Aether plugins are designed to be:
//
//   - Simple to implement
//   - Safe by design
//   - Flexible enough for enterprise extension
//   - Stable across Aether versions
//
// This package provides only interface definitions and documentation.
// Registration logic, routing, and execution glue live in the aether
// package and internal/plugin_registry components.
package plugins
