# Aether

> **Aether** is a legal, robots.txt-compliant, open‑data retrieval and normalization toolkit for Go — built for **LLM / RAG / agentic AI systems**.

Aether turns arbitrary public web content into **structured, LLM‑ready representations** (JSON + TOON), with strong guarantees around **legality**, **robots.txt compliance**, **caching**, and **predictable output schemas**.

- ✅ **Pure Go library** (`import "github.com/Nibir1/Aether/aether"`)
- ✅ **Robots.txt‑compliant HTTP client** with per‑host fairness and optional host-level robots override
- ✅ **Multi‑layer cache** (memory + file + redis via composite cache)
- ✅ **Article extraction**, **RSS/Atom** parsing, **OpenAPI** connectors
- ✅ **Plugins** (Source / Transform / Display)
- ✅ **Canonical JSON** + **TOON 2.0** + **Lite TOON** + **BTON (binary)**
- ✅ **Streaming outputs** (JSONL + TOON event streams)
- ✅ Fully tested across **Normal & Robots Override** modes
- ✅ Designed for **AI engineers**, **backend devs**, and **agent frameworks**

---

## Table of Contents

1. [Why Aether?](#why-aether)
2. [Feature Overview](#feature-overview)
3. [Architecture](#architecture)
4. [Installation](#installation)
5. [Quickstart](#quickstart)
6. [Usage by Feature](#usage-by-feature)
   - [Search & Normalize](#1-search--normalize)
   - [HTTP Fetch & Detect](#2-http-fetch--detect)
   - [HTML Parsing & Article Extraction](#3-html-parsing--article-extraction)
   - [RSS / Atom Feeds](#4-rss--atom-feeds)
   - [OpenAPI Integrations](#5-openapi-integrations)
   - [Crawling](#6-crawling)
   - [Batch Fetch](#7-batch-fetch)
   - [SmartQuery Routing](#8-smartquery-routing)
   - [Display & Markdown Rendering](#9-display--markdown-rendering)
   - [Plugins (Source / Transform / Display)](#10-plugins-source--transform--display)
   - [JSONL Streaming](#11-jsonl-streaming)
   - [TOON, Lite TOON & BTON](#12-toon-lite-toon--bton)
   - [TOON Streaming](#13-toon-streaming)
   - [Error Handling](#14-error-handling)
   - [Configuration & Caching](#15-configuration--caching)
   - [Robots Override](#16-robots--override)
7. [cmd/ Test Programs](#cmd-test-programs)
8. [Status & Roadmap](#status--roadmap)
9. [License](#license)

---

## Why Aether?

Modern AI / LLM systems need **structured, legal access** to the public web. Most options fall into two extremes:

- 🔒 **Paid search APIs / proprietary services**
  - Expensive
  - Rate‑limited
  - Closed schemas
  - Often not robots‑aware from your perspective

- 🧪 **Ad‑hoc scraping scripts**
  - Legally risky
  - No robots.txt handling
  - Brittle HTML parsing
  - No unified schema, no caching discipline

**Aether** sits between these extremes:

- **Legal by design**
  - All HTTP calls are made via a **robots.txt‑compliant** internal client.
  - Per‑host throttling, polite concurrency, and no captcha/anti‑bot bypassing.
- **Structured by default**
  - Everything ends up as a **canonical `model.Document`** and optionally **TOON 2.0**.
- **LLM‑first**
  - Output formats are designed for **RAG**, **embedding pipelines**, and **agentic workflows**.
- **Extensible**
  - Plugin architecture for custom data sources, enrichers, and renderers.
- **Cost‑effective**
  - Aether itself is a **free, open‑source** alternative to many paid “web search / content extraction” APIs.
  - You pay only for your own infrastructure + outbound bandwidth.

Aether is ideal if you are:

- Building **RAG systems** that need legal web context
- Implementing **agent frameworks** that must “read the web”
- Building **LLM tooling** for research, journalism, or public‑data analytics
- Looking for a **self‑hosted alternative** to paid web retrieval APIs

---

## Feature Overview

### Core

- **Robots‑aware HTTP client**
  - `Client.Fetch`, `Client.FetchRaw`, `Client.FetchText`, `Client.FetchJSON`
- **Composite Caching**
  - Memory + file + redis via `internal/cache`, configurable via `Config`
- **Search Pipeline**
  - `Client.Search` → `SearchResult` → `NormalizeSearchResult`
- **Normalization**
  - Everything converges into `internal/model.Document` (exported as `aether.NormalizedDocument`)
  - Supports articles, feeds, text, JSON, entities, and plugin outputs

### Content Understanding

- **Detect**
  - `Client.Detect` — MIME + charset + HTML metadata
- **HTML Parsing**
  - `Client.ParseHTML` — headings, paragraphs, links, meta
- **Article Extraction**
  - `Client.ExtractArticleFromHTML` / `Client.ExtractArticle`

### Feeds & APIs

- **RSS / Atom**
  - `Client.FetchRSS` / `Client.ParseRSS`
- **OpenAPI Modules**
  - `WikipediaSummary`
  - `HackerNewsTopStories`
  - `GitHubReadme`
  - `WhiteHouseRecentPosts`
  - `GovernmentPress`
  - `WeatherAt` (MET Norway)
  - `WikidataLookup`

### Higher‑Level Retrieval

- **Crawl**
  - Depth‑limited, robots‑aware, polite crawling
  - `Client.Crawl` with `CrawlOptions`
- **Batch**
  - Concurrent multi‑URL fetch
  - `Client.Batch`

### LLM‑Friendly Output

- **Canonical JSON**
  - `NormalizeSearchResult` + `MarshalSearchResultJSON`
- **TOON 2.0**
  - `ToTOON`, `MarshalTOON`, `MarshalTOONPretty`
- **Lite TOON**
  - `MarshalTOONLite`, `MarshalTOONLitePretty`
- **BTON Binary**
  - `MarshalBTON`, `UnmarshalBTON`, `MarshalBTONFromModel`
- **JSONL Streaming**
  - `StreamNormalizedJSONL`, `StreamSearchResultJSONL`, `StreamFeedJSONL`
- **TOON Streaming**
  - `StreamTOON`, `StreamSearchResultTOON`

### Plugins

- **Source Plugins**
  - Custom legal/public data sources (e.g., HN plugin, custom JSON API plugin)
- **Transform Plugins**
  - Post‑normalization enrichment (summarization, entity extraction, metadata)
- **Display Plugins**
  - Render normalized documents as HTML, ANSI, PDF, etc.
- **Registry**
  - Thread‑safe plugin registration & lookup via `plugins.Registry` and `Client.Register*Plugin`

### Display

- **Markdown display**
  - `RenderMarkdown`, `RenderMarkdownWithTheme`
- **Preview**
  - `RenderPreview`, `RenderPreviewWithTheme`
- **Tables**
  - `RenderTable`, `RenderTableWithTheme`
- **Unified Render**
  - `Render(ctx, format, doc)` and `RenderSearchResult(ctx, format, sr)` (built‑in + display plugins)

---

## Architecture

Aether is centered around the **Client** type:

```go
cli, err := aether.NewClient(
    aether.WithDebugLogging(false),
)
if err != nil {
    // handle error
}
```

### High‑Level Flow

1. **Input**: URL or text query
2. **SmartQuery** (optional): classify intent and routing
3. **Search / Fetch / OpenAPI / Plugins**: get raw content
4. **Detect / Extract / RSS**: parse & understand structure
5. **Normalize**: produce `model.Document`
6. **Transforms**: apply TransformPlugins
7. **Output**:
   - JSON (canonical)
   - TOON (full / lite / BTON)
   - Markdown / Preview / Tables
   - JSONL / TOON streams
   - DisplayPlugin formats (HTML, PDF, ANSI, …)

### Detailed Architecture Diagram

```text
                               ┌───────────────────────────┐
                               │         Your App          │
                               │  (LLM / RAG / Agent / UI) │
                               └─────────────┬─────────────┘
                                             │
                                      Aether Client
                                             │
      ┌──────────────────────────────────────┼──────────────────────────────────────┐
      │                                      │                                      │
      ▼                                      ▼                                      ▼
┌───────────────┐                    ┌───────────────┐                       ┌─────────────┐
│  SmartQuery   │                    │  Direct Fetch │                       │   Plugins   │
│  (intent,     │                    │  & OpenAPIs   │                       │  (Source)   │
│  routing)     │                    │               │                       │             │
└──────┬────────┘                    └──────┬────────┘                       └──────┬──────┘
       │                                    │                                       │
       │                         ┌──────────▼─────────┐                             │
       │                         │ robots-aware HTTP  │                             │
       │                         │  + Composite Cache │                             │
       │                         │ (memory/file/redis)│                             │
       │                         └─────────┬──────────┘                             │
       │                                   │                                        │
       │                         ┌─────────▼───────────┐                            │
       │                         │  Content Detection  │                            │
       │                         │  (MIME, charset,    │                            │
       │                         │   HTML meta)        │                            │
       │                         └─────────┬───────────┘                            │
       │                                   │                                        │
       │              ┌────────────────────┼─────────────────────┐                  │
       │              ▼                    ▼                     ▼                  │
       │       ┌──────────────┐    ┌───────────────┐     ┌──────────────┐           │
       │       │ HTML Parser  │    │  RSS / Atom   │     │ OpenAPI      │           │
       │       │ + Extractor  │    │  Parser       │     │ (Wikipedia,  │           │
       │       │ (Article)    │    │               │     │  HN, etc.)   │           │
       │       └──────┬───────┘    └──────┬────────┘     └──────┬───────┘           │
       │              │                   │                     │                   │
       └──────────────┼───────────────────┼─────────────────────┼───────────────────┘
                      ▼                   ▼                     ▼
                ┌────────────────────────────────────────────────────────────┐
                │                   Normalization Layer                      │
                │             (internal/normalize → model.Document)          │
                │  • SearchDocument → Document                               │
                │  • Article → Sections (body, summary)                      │
                │  • Feed → Sections (feed_item)                             │
                │  • Entities → Sections (entity)                            │
                └───────────────┬────────────────────────────────────────────┘
                                │
                                ▼
                      ┌─────────────────────┐
                      │ Transform Plugins   │
                      │ (plugins.Transform) │
                      │  – summarization    │
                      │  – keyword/meta     │
                      │  – enrichment       │
                      └─────────┬───────────┘
                                │
                                ▼
          ┌────────────────────────────────────────────────────────────────┐
          │                         Output Layer                           │
          │                                                                │
          │  JSON / TOON Core:                                             │
          │    • MarshalSearchResultJSON                                   │
          │    • ToTOON / MarshalTOON / Lite / BTON                        │
          │    • StreamNormalizedJSONL / StreamSearchResultJSONL           │
          │    • StreamTOON / StreamSearchResultTOON                       │
          │                                                                │
          │  Display:                                                      │
          │    • RenderMarkdown / Preview / Tables                         │
          │    • Render(ctx, format, doc) (built-in + DisplayPlugins)      │
          │                                                                │
          │  Plugins (Display):                                            │
          │    • HTML / ANSI / PDF / custom formats                        │
          └────────────────────────────────────────────────────────────────┘
```

---

## Installation

```bash
go get github.com/Nibir1/Aether@v1.0.0
```

Then import:

```go
import "github.com/Nibir1/Aether/aether"
```

---

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Nibir1/Aether/aether"
)

func main() {
    cli, err := aether.NewClient(
        aether.WithDebugLogging(false),
    )
    if err != nil {
        log.Fatalf("failed to create Aether client: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    // Simple factual lookup
    sr, err := cli.Search(ctx, "Finland")
    if err != nil {
        log.Fatalf("Search error: %v", err)
    }

    // Normalize → JSON
    norm := cli.NormalizeSearchResult(sr)
    jsonBytes, err := cli.MarshalSearchResultJSON(sr)
    if err != nil {
        log.Fatalf("Marshal JSON error: %v", err)
    }

    fmt.Println("Title:", norm.Title)
    fmt.Println("Excerpt:", norm.Excerpt)
    fmt.Println("JSON:", string(jsonBytes))
}
```

---

## Usage by Feature

### 1. Search & Normalize

Use `Client.Search` to handle both URL and text queries. Aether routes internally (plugins, Wikipedia fallback, direct fetch).

```go
ctx := context.Background()

sr, err := cli.Search(ctx, "Helsinki weather")
if err != nil {
    log.Fatal(err)
}

// Inspect search plan
fmt.Println("Intent:", sr.Plan.Intent)
fmt.Println("Source:", sr.Plan.Source)

// Normalize to canonical model.Document
norm := cli.NormalizeSearchResult(sr)
fmt.Println("Normalized kind:", norm.Kind)
fmt.Println("Title:", norm.Title)
```

---

### 2. HTTP Fetch & Detect

#### Basic fetch

```go
res, err := cli.Fetch(ctx, "https://example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Status:", res.StatusCode)
fmt.Println("Body size:", len(res.Body))
```

#### FetchRaw / FetchText / FetchJSON

```go
// Raw bytes
body, hdr, err := cli.FetchRaw(ctx, "https://example.com/data.json")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Content-Type:", hdr.Get("Content-Type"))
fmt.Println("Raw length:", len(body))

// Text
text, _, err := cli.FetchText(ctx, "https://example.com/README.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println("README snippet:", text[:200])

// JSON into struct
var payload struct {
    Name string `json:"name"`
}
if err := cli.FetchJSON(ctx, "https://example.com/api/info", &payload); err != nil {
    log.Fatal(err)
}
fmt.Println("Name:", payload.Name)
```

#### Detect content type & HTML metadata

```go
det, err := cli.Detect(ctx, "https://example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println("MIME:", det.MIME)
fmt.Println("IsBinary:", det.IsBinary)
fmt.Println("Title:", det.Title)
fmt.Println("Canonical URL:", det.Canonical)
```

---

### 3. HTML Parsing & Article Extraction

```go
res, err := cli.Fetch(ctx, "https://example.com/article.html")
if err != nil {
    log.Fatal(err)
}

// Parse HTML structure
parsed, err := cli.ParseHTML(res.Body)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Page Title:", parsed.Title)
for _, h := range parsed.Headings {
    fmt.Printf("H%d: %s\n", h.Level, h.Text)
}

// Extract main article content (readability-style)
art, err := cli.ExtractArticle(ctx, "https://example.com/article.html")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Article Title:", art.Title)
fmt.Println("Byline:", art.Byline)
fmt.Println("Excerpt:", art.Excerpt)
fmt.Println("First 300 chars:", art.Content[:300])
```

---

### 4. RSS / Atom Feeds

```go
feed, err := cli.FetchRSS(ctx, "https://example.com/feed.xml")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Feed title:", feed.Title)
for _, item := range feed.Items {
    fmt.Println("-", item.Title, "→", item.Link)
}
```

You can also stream feed items as JSONL:

```go
if err := cli.StreamFeedJSONL(ctx, os.Stdout, feed); err != nil {
    log.Fatal(err)
}
```

---

### 5. OpenAPI Integrations

#### Wikipedia

```go
sum, err := cli.WikipediaSummary(ctx, "Finland")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Title:", sum.Title)
fmt.Println("Description:", sum.Description)
fmt.Println("Extract:", sum.Extract)
```

#### Hacker News

```go
stories, err := cli.HackerNewsTopStories(ctx, 5)
if err != nil {
    log.Fatal(err)
}
for _, s := range stories {
    fmt.Printf("[%d] %s (%s)\n", s.Score, s.Title, s.URL)
}
```

#### GitHub README

```go
readme, err := cli.GitHubReadme(ctx, "golang", "go", "master")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Repo:", readme.Owner+"/"+readme.Repo)
fmt.Println("README excerpt:", readme.Content[:300])
```

#### Government Press / White House / Weather / Wikidata

```go
// White House
posts, _ := cli.WhiteHouseRecentPosts(ctx, 3)
// Government press
press, _ := cli.GovernmentPress(ctx, 5)
// Weather via MET Norway
weather, _ := cli.WeatherAt(ctx, 60.1699, 24.9384, 12) // Helsinki approx
// Wikidata
ent, _ := cli.WikidataLookup(ctx, "Helsinki")
```

---

### 6. Crawling

```go
opts := aether.CrawlOptions{
    MaxDepth:       1,
    MaxPages:       10,
    SameHostOnly:   true,
    FetchDelay:     500 * time.Millisecond,
    AllowedDomains: []string{"example.com"},
    Visitor: aether.CrawlVisitorFunc(func(ctx context.Context, p *aether.CrawledPage) error {
        fmt.Println("Visited:", p.URL, "Status:", p.StatusCode)
        fmt.Println("Excerpt:", p.Content[:200])
        return nil
    }),
}

if err := cli.Crawl(ctx, "https://example.com", opts); err != nil {
    log.Fatal(err)
}
```

---

### 7. Batch Fetch

```go
urls := []string{
    "https://example.com",
    "https://example.org",
    "https://example.net",
}

res, err := cli.Batch(ctx, urls, aether.BatchOptions{
    Concurrency: 4,
})
if err != nil {
    log.Fatal(err)
}

for _, item := range res.Results {
    if item.Err != nil {
        fmt.Println("Error fetching", item.URL, ":", item.Err)
        continue
    }
    fmt.Println("Fetched", item.URL, "status", item.StatusCode, "bytes", len(item.Body))
}
```

---

### 8. SmartQuery Routing

```go
plan := cli.SmartQuery("latest hacker news about go generics")
fmt.Println("Intent:", plan.Intent)
fmt.Println("IsQuestion:", plan.IsQuestion)
fmt.Println("HasURL:", plan.HasURL)
fmt.Println("PrimarySources:", plan.PrimarySources)
fmt.Println("UseOpenAPIs:", plan.UseOpenAPIs)
fmt.Println("UsePlugins:", plan.UsePlugins)
```

You can use this to drive higher‑level agent decisions before calling `Search`.

---

### 9. Display & Markdown Rendering

Render normalized documents for CLIs, logs, or debugging:

```go
sr, err := cli.Search(ctx, "Finland")
if err != nil {
    log.Fatal(err)
}

norm := cli.NormalizeSearchResult(sr)

// Markdown
md := cli.RenderMarkdown(norm)
fmt.Println(md)

// Preview
preview := cli.RenderPreview(norm)
fmt.Println(preview)

// Table
table := cli.RenderTable(
    []string{"Title", "URL"},
    [][]string{
        {norm.Title, norm.Metadata["page_url"]},
    },
)
fmt.Println(table)
```

Render via the **unified dispatcher**, including DisplayPlugins:

```go
out, err := cli.Render(ctx, "markdown", norm) // or "preview", or plugin formats like "html"
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(out))
```

---

### 10. Plugins (Source / Transform / Display)

#### Source Plugins

Implement `plugins.SourcePlugin`:

```go
type MySourcePlugin struct{}

func (p *MySourcePlugin) Name() string         { return "my_source" }
func (p *MySourcePlugin) Description() string  { return "Custom source plugin example" }
func (p *MySourcePlugin) Capabilities() []string {
    return []string{"custom", "example"}
}

func (p *MySourcePlugin) Fetch(ctx context.Context, query string) (*plugins.Document, error) {
    return &plugins.Document{
        Source:  "plugin:my_source",
        Kind:    plugins.DocumentKindText,
        Title:   "Result for " + query,
        Excerpt: "Custom plugin result",
        Content: "Full body from MySourcePlugin",
    }, nil
}
```

Register it on the client:

```go
mySrc := &MySourcePlugin{}
if err := cli.RegisterSourcePlugin(mySrc); err != nil {
    log.Fatal(err)
}
```

Aether’s `Search` will now be able to route queries through your plugin.

#### Transform Plugins

Transform normalized documents (via `NormalizeSearchResult`):

```go
type MyTransform struct{}

func (t *MyTransform) Name() string        { return "my_transform" }
func (t *MyTransform) Description() string { return "Adds a custom metadata flag" }

func (t *MyTransform) Apply(ctx context.Context, doc *plugins.Document) (*plugins.Document, error) {
    if doc.Metadata == nil {
        doc.Metadata = map[string]string{}
    }
    doc.Metadata["my_transform.applied"] = "true"
    return doc, nil
}
```

Register:

```go
if err := cli.RegisterTransformPlugin(&MyTransform{}); err != nil {
    log.Fatal(err)
}

// Any NormalizeSearchResult(...) call now passes through MyTransform
norm := cli.NormalizeSearchResult(sr)
fmt.Println(norm.Metadata["my_transform.applied"]) // "true"
```

#### Display Plugins

Render normalized docs into custom formats (HTML, PDF, ANSI, etc.):

```go
type MyHTMLDisplay struct{}

func (d *MyHTMLDisplay) Name() string        { return "my_html" }
func (d *MyHTMLDisplay) Description() string { return "Simple HTML renderer" }
func (d *MyHTMLDisplay) Format() string      { return "html" }

func (d *MyHTMLDisplay) Render(ctx context.Context, doc *plugins.Document) ([]byte, error) {
    html := "<html><head><title>" + doc.Title + "</title></head><body>"
    html += "<h1>" + doc.Title + "</h1>"
    html += "<p>" + doc.Excerpt + "</p>"
    html += "<pre>" + doc.Content + "</pre>"
    html += "</body></html>"
    return []byte(html), nil
}
```

Register & use:

```go
if err := cli.RegisterDisplayPlugin(&MyHTMLDisplay{}); err != nil {
    log.Fatal(err)
}

norm := cli.NormalizeSearchResult(sr)
htmlBytes, err := cli.Render(ctx, "html", norm)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(htmlBytes))
```

---

### 11. JSONL Streaming

#### NormalizedDocument → JSONL

```go
norm := cli.NormalizeSearchResult(sr)

if err := cli.StreamNormalizedJSONL(ctx, os.Stdout, norm); err != nil {
    log.Fatal(err)
}
```

Output looks like:

```jsonl
{"type":"document","data":{...}}
{"type":"metadata","data":{...}}
{"type":"section","data":{...}}
{"type":"section","data":{...}}
...
```

#### SearchResult → JSONL

```go
if err := cli.StreamSearchResultJSONL(ctx, os.Stdout, sr); err != nil {
    log.Fatal(err)
}
```

#### Feed → JSONL

```go
feed, err := cli.FetchRSS(ctx, "https://example.com/feed.xml")
if err != nil {
    log.Fatal(err)
}

if err := cli.StreamFeedJSONL(ctx, os.Stdout, feed); err != nil {
    log.Fatal(err)
}
```

---

### 12. TOON, Lite TOON & BTON

#### TOON from SearchResult

```go
tdoc := cli.ToTOON(sr)
b, err := cli.MarshalTOON(sr)
if err != nil {
    log.Fatal(err)
}
fmt.Println("TOON JSON:", string(b))
```

#### Pretty TOON

```go
pretty, _ := cli.MarshalTOONPretty(sr)
fmt.Println(string(pretty))
```

#### Lite TOON (compact JSON)

```go
lite, _ := cli.MarshalTOONLite(sr)
fmt.Println("Lite TOON:", string(lite))
```

#### BTON (binary TOON)

```go
btonBytes, err := cli.MarshalBTON(sr)
if err != nil {
    log.Fatal(err)
}

// Later / elsewhere:
tdoc2, err := cli.UnmarshalBTON(btonBytes)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Decoded TOON kind:", tdoc2.Kind)
```

---

### 13. TOON Streaming

Stream TOON as **JSONL events** (doc_start/doc_meta/token/doc_end):

```go
norm := cli.NormalizeSearchResult(sr)

// NormalizedDocument → TOON stream
if err := cli.StreamTOON(ctx, os.Stdout, norm); err != nil {
    log.Fatal(err)
}
```

Or directly from a SearchResult:

```go
if err := cli.StreamSearchResultTOON(ctx, os.Stdout, sr); err != nil {
    log.Fatal(err)
}
```

Example output (JSONL):

```jsonl
{"event":"doc_start","kind":"article","source_url":"https://...","title":"...","excerpt":"..."}
{"event":"doc_meta","attrs":{"aether.intent":"lookup"}}
{"event":"token","token":{"type":"heading","category":"content","text":"...","attrs":{"level":"1"}}}
{"event":"token","token":{"type":"text","category":"content","text":"..."}} 
{"event":"doc_end"}
```

This is ideal for **agent streaming**, **indexing pipelines**, or **RAG pre‑processing**.

---

### 14. Error Handling

Aether exposes structured error kinds:

```go
import "errors"

res, err := cli.Fetch(ctx, "https://example.com")
if err != nil {
    var ae *aether.Error
    if errors.As(err, &ae) {
        switch ae.Kind {
        case aether.ErrorKindRobots:
            fmt.Println("Robots.txt blocked the request:", ae.Msg)
        case aether.ErrorKindHTTP:
            fmt.Println("HTTP error:", ae.Msg)
        default:
            fmt.Println("Aether error:", ae.Msg)
        }
    } else {
        fmt.Println("Generic error:", err)
    }
}
```

Error kinds include:

- `ErrorKindUnknown`
- `ErrorKindConfig`
- `ErrorKindHTTP`
- `ErrorKindRobots`
- `ErrorKindParsing`

---

### 15. Configuration & Caching

`NewClient` accepts option functions that modify internal `config.Config`:

```go
cli, err := aether.NewClient(
    aether.WithUserAgent("MyApp/1.0 (+https://example.com)"),
    aether.WithRequestTimeout(10*time.Second),
    aether.WithConcurrency(16, 4), // 16 hosts, 4 per host
    aether.WithDebugLogging(false),
)
if err != nil {
    log.Fatal(err)
}
```

Inspect effective configuration:

```go
cfg := cli.EffectiveConfig()
fmt.Println("UA:", cfg.UserAgent)
fmt.Println("Request timeout:", cfg.RequestTimeout)
fmt.Println("Memory cache enabled:", cfg.EnableMemoryCache)
fmt.Println("File cache dir:", cfg.CacheDirectory)
```

The internal composite cache supports:

- In‑memory LRU
- File‑backed disk cache
- Redis cache (for shared/distributed setups)

Configuration is wired through `internal/config` + `internal/cache` and surfaced via `EffectiveConfig`.

---

### 16. Robots Override

Aether `NewClient` supports **robots override** options to selectively bypass `robots.txt` for certain hosts. This is **host-specific** and still respects robots rules for all other domains.

#### Usage: Boolean Enable (Global Override)

Enable robots override for advanced use-cases (use with caution, responsibility lies with the caller):

```go
cli, err := aether.NewClient(
    aether.WithDebugLogging(true),
    aether.WithRobotsOverride(true), // enables global override mode
)
if err != nil {
    log.Fatal(err)
}
```

#### Usage: Host List (Selective Override)

Override robots rules for specific hosts only:

```go
cli, err := aether.NewClient(
    aether.WithDebugLogging(true),
    aether.WithRobotsOverride(
        "hnrss.org",
        "news.ycombinator.com",
        "example.com",
    ),
)
if err != nil {
    log.Fatal(err)
}
```

#### Inspect Configuration

Check which hosts are allowed and whether override is enabled:

```go
cfg := cli.EffectiveConfig()
fmt.Println("Robots Override Enabled:", cfg.RobotsOverrideEnabled)
fmt.Println("Robots Allowed Hosts:", cfg.RobotsAllowedHosts)
```

#### Notes

* Hosts are matched **case-insensitively** and without port.
* Responsibility for ignoring robots rules lies entirely with the caller.
* Aether will still obey robots rules for all hosts not explicitly listed.
* Useful for internal testing, public data aggregation, or legal-use scenarios where host consent is verified.

---

## cmd/ Test Programs

The repository includes several **executable test programs** under `cmd/` for manual testing and examples.

Typical layout (may evolve):

```text
cmd/
  test_batch/
  test_cache/
  test_crawl/
  test_display_plugins/
  test_fetch/
  test_jsonl/
  test_normalize_merge/
  test_openapi/
  test_plugins/
  test_rss/
  test_search_display/
  test_smartquery/
  test_toon_lite_bton/
  test_toon_stream/
  test_transforms/
```

Run any test with:

```bash
go run ./cmd/test_fetch
go run ./cmd/test_search_display
go run ./cmd/test_openapi
# etc.
```

Each test is a small `main.go` that exercises a specific subsystem:

- `test_fetch` — robots‑aware HTTP fetch + detect
- `test_search_display` — high‑level Search + Normalize + Display
- `test_openapi` — Wikipedia, HN, GitHub, GovPress, Weather, Wikidata
- `test_rss` — RSS/Atom fetch & parse
- `test_crawl` — crawl API
- `test_batch` — batch concurrent fetch
- `test_jsonl` — JSONL streaming
- `test_toon_stream` — TOON event streaming
- `test_toon_lite_bton` — TOON Lite + BTON encode/decode
- `test_plugins` — plugin registration and wiring
- `test_transforms` — TransformPlugins pipeline
- `test_display_plugins` — DisplayPlugin routing
- `test_smartquery` — SmartQuery classification
- `test_cache` — cache behavior & config
- `test_normalize_merge` — normalization/merge invariants

These are great **reference implementations** when integrating Aether into your own app.

---

## Status & Roadmap

Aether’s original roadmap is approximately **90% complete**:

- ✅ Core HTTP + robots.txt + caching
- ✅ Detect / HTML / article extraction
- ✅ RSS/Atom subsystem
- ✅ OpenAPI integrations
- ✅ Search pipeline + SmartQuery
- ✅ Normalization → `model.Document`
- ✅ TOON 2.0 + Lite + BTON
- ✅ JSONL & TOON streaming
- ✅ Plugin system (Source / Transform / Display)
- ✅ Display subsystem (Markdown, preview, tables)
- ✅ Manual tests (`cmd/test_*`)

Planned / nice‑to‑have improvements:

- 🔸 More OpenAPI integrations (free, public)
- 🔸 More powerful SmartQuery routing / ranking
- 🔸 Additional DisplayPlugins (HTML templates, ANSI themes, PDF)
- 🔸 Richer TransformPlugins (summarization, entity extraction, auto‑tagging)
- 🔸 Higher‑level convenience helpers for common AI workflows

Despite being early, Aether already forms a **production‑ready foundation** for **LLM‑aware web retrieval pipelines**.

---

## 📄 License

Aether is released under the **MIT License** (see `LICENSE`).  

## 🤝 Contributing

Ideas, issues, and PRs are welcome:

1. Fork the repo
2. Create a feature branch
3. Open a PR with a clear description and demo steps

---

## 👤 Developer Spotlight

**Nahasat Nibir** — Building intelligent, high‑performance developer tools and AI‑powered systems in Go and Python.

- GitHub: https://github.com/Nibir1
- LinkedIn: https://www.linkedin.com/in/nibir-1/
- ArtStation: https://www.artstation.com/nibir

---

<div align="center">
Aether is a legal, robots.txt-compliant, open‑data retrieval and normalization toolkit 
<br />
- Built for LLM / RAG / Agentic AI Systems -
<br />
<a href="https://github.com/Nibir1/Helix/issues">🐞 Report Bug</a> ·
<a href="https://github.com/Nibir1/Helix/issues">💡 Request Feature</a> ·
⭐ Star the project
</div>