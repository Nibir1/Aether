# Custom API Source Plugin (Example)

This directory contains a fully functional example of how to build a
SourcePlugin for **any public or internal JSON API** using Aether's
legal, robots.txt-compliant HTTP pipeline.

This plugin is designed as a **template** for developers integrating:

-   Internal microservices
-   Public JSON datasets (weather, transport, economic indicators, etc.)
-   University research APIs
-   Open data portals
-   Low-complexity REST APIs
-   Company-internal knowledge services

It is intentionally written to be realistic, production‑quality, and
easy to adapt.

------------------------------------------------------------------------

## 🚀 What This Plugin Does

-   Accepts a natural-language **query** string\
-   Calls a fictional API endpoint:

```{=html}
<!-- -->
```
    GET https://api.example.com/v1/info?query=<term>

-   Parses JSON into Go structs\
-   Converts the API result into a structured `plugins.Document`
-   Adds sections, metadata, content, and excerpt\
-   Returns the result to Aether's normalization and display layers

This plugin is a **canonical example** of how to write data-driven
plugins for Aether.

------------------------------------------------------------------------

## 📦 Files

    plugin.go      — Full example implementation
    readme.md      — Documentation (this file)

------------------------------------------------------------------------

## 🧩 Registering the Plugin

``` go
package main

import (
    "context"
    "fmt"

    "github.com/Nibir1/Aether/aether"
    "github.com/Nibir1/Aether/plugins/examples/custom_api_plugin"
)

func main() {
    cli, _ := aether.NewClient()

    // Create plugin instance
    custom := custom_api_plugin.New(cli)

    // Register plugin
    if err := cli.RegisterSourcePlugin(custom); err != nil {
        panic(err)
    }

    // Execute plugin query
    doc, err := custom.Fetch(context.Background(), "climate change")
    if err != nil {
        panic(err)
    }

    fmt.Println("Document Title:", doc.Title)
}
```

------------------------------------------------------------------------

## 🧠 JSON Format Expected from API

This plugin expects the external API to return JSON structured like:

``` json
{
  "title": "Sample Title",
  "excerpt": "Short summary",
  "content": "Long-form article or text",
  "metadata": {
    "source": "example",
    "rating": "A+"
  },
  "sections": [
    {
      "heading": "Overview",
      "text": "paragraphs...",
      "meta": { "version": "1" }
    }
  ]
}
```

This models many real-world APIs deployed in companies and public
datasets.

------------------------------------------------------------------------

## 📝 How It Works

### 1. Build the request URL

The query is URL‑encoded and added as `?query=<term>`.

### 2. Fetch using Aether

Aether handles:

-   robots.txt compliance\
-   rate limiting\
-   retries\
-   caching\
-   decompression\
-   error wrapping

This keeps your plugin **legal, safe, and stable**.

### 3. Parse JSON

Data is decoded into:

``` go
type apiResponse struct {
    Title    string
    Excerpt  string
    Content  string
    Metadata map[string]string
    Sections []apiResponseSection
}
```

### 4. Convert to plugins.Document

The plugin builds a structured document with:

-   Title\
-   Excerpt\
-   Content\
-   Metadata\
-   Multiple Sections

Aether later normalizes this into JSON or TOON format.

------------------------------------------------------------------------

## 🧭 Capabilities

The plugin advertises:

``` go
[]string{"data", "info", "custom", "api"}
```

SmartQuery routing can now auto‑select this plugin for general info/data
queries.

------------------------------------------------------------------------

## 🔐 Legal & Safety Notes

-   No scraping is performed\
-   Robots.txt is automatically respected\
-   The plugin cannot bypass anti-bot systems\
-   Only legal, public data sources may be used\
-   Internal APIs must be authorized and permitted

Plugins must respect **Aether's strict safety rules**.

------------------------------------------------------------------------

## 🛠 Extend This Plugin

You can extend or adapt this example by:

-   Adding authentication headers\
-   Adding POST requests for advanced APIs\
-   Parsing more complex section structures\
-   Adding TransformPlugins for summarization\
-   Creating DisplayPlugins for beautiful rendering

------------------------------------------------------------------------

## 📄 License

This example is part of the Aether open-source project and is freely
available for learning, adaptation, and reuse.
