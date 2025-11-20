# Hacker News Source Plugin (Example)

This directory contains a complete example implementation of an Aether
**SourcePlugin** that retrieves top stories from **Hacker News** using
the public Firebase API.

The Hacker News API is:

-   **Free**
-   **Public**
-   **Robots.txt-friendly**
-   **No API key required**
-   **Suitable for safe use inside Aether plugins**

This example demonstrates how to implement a well-behaved plugin that
respects Aether's legal responsibilities while providing real, useful
data.

------------------------------------------------------------------------

## 🚀 What This Plugin Does

-   Fetches the list of **top story IDs**
-   Retrieves each individual story's JSON metadata
-   Builds a `plugins.Document` of kind `feed`
-   Adds each story as a structured `Section`
-   Includes metadata such as score, author, timestamp, and URL
-   Can be registered with any `*aether.Client`

The output integrates seamlessly into Aether's:

-   Normalization system\
-   Display subsystem (Markdown, TOON, etc.)\
-   SmartQuery routing (via plugin capabilities)

------------------------------------------------------------------------

## 📦 Files

    hn.go          — Full SourcePlugin implementation
    readme.md      — Documentation (this file)

------------------------------------------------------------------------

## 🧩 Registering the Plugin

``` go
package main

import (
    "context"
    "fmt"

    "github.com/Nibir1/Aether/aether"
    "github.com/Nibir1/Aether/plugins/examples/hn_plugin"
)

func main() {
    cli, _ := aether.NewClient()

    // Create plugin instance
    hn := hn_plugin.New(cli, 10) // fetch top 10 stories

    // Register plugin
    if err := cli.RegisterSourcePlugin(hn); err != nil {
        panic(err)
    }

    // Use SmartQuery or direct execution
    doc, err := hn.Fetch(context.Background(), "")
    if err != nil {
        panic(err)
    }

    fmt.Println("HN Document Title:", doc.Title)
}
```

------------------------------------------------------------------------

## 📝 How It Works

### 1. Fetch Top Stories

The plugin asks Aether to fetch:

    https://hacker-news.firebaseio.com/v0/topstories.json

Aether handles:

-   robots.txt
-   caching
-   gzip
-   retries
-   error wrapping
-   concurrency limits

### 2. Fetch Story Metadata

For each ID:

    https://hacker-news.firebaseio.com/v0/item/<ID>.json

Parsed into:

``` go
type hnStory struct {
    ID    int64
    Title string
    URL   string
    By    string
    Score int
    Time  int64
    Kids  []int64
    Type  string
}
```

### 3. Convert to Plugin Document

Built as:

``` go
&plugins.Document{
    Source: "plugin:hackernews",
    Kind:   plugins.DocumentKindFeed,
    Sections: []plugins.Section{
        {
            Role:  "feed_item",
            Title: story.Title,
            Text:  "<snippet>",
            Meta:  map[string]string{...},
        },
    },
}
```

Aether later converts this into its full normalized document model.

------------------------------------------------------------------------

## 🧭 Capabilities

The plugin advertises:

``` go
[]string{"news", "tech", "hn", "hackernews"}
```

This allows Aether's SmartQuery engine to route tech/news queries
automatically to this plugin.

------------------------------------------------------------------------

## 🔐 Legal & Safety Notes

-   Hacker News API explicitly allows programmatic access\
-   No authentication required\
-   Follows robots.txt\
-   No scraping of HTML pages\
-   Only uses documented public endpoints

This plugin is 100% compliant with Aether's legal & ethical rules.

------------------------------------------------------------------------

## 🛠 Extend This Plugin

You can enhance it by:

-   Adding filtering (e.g., top stories in "Ask HN" only)
-   Adding support for "beststories", "newstories", etc.
-   Including comment counts (`Kids` length)
-   Adding a TransformPlugin that summarizes all stories
-   Adding a DisplayPlugin that renders ANSI-colored HN feed

------------------------------------------------------------------------

## 📄 License

This example is provided as part of the Aether project and is intended
for public use, learning, and extension.
