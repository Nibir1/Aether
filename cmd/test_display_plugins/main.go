// cmd/test_display_plugins/main.go
//
// Test for Aether's DisplayPlugin system (Strict Mode).
// This test registers a dummy HTML display plugin, performs a real
// Search("Finland"), and renders the normalized SearchResult using the
// custom plugin. It also confirms strict-mode error behavior when attempting
// to render an unregistered format (e.g. "pdf").
//

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

//
// ───────────────────────────────────────────────
//                Dummy HTML Plugin
// ───────────────────────────────────────────────
//

// HTMLPlugin implements plugins.DisplayPlugin.
type HTMLPlugin struct{}

func (HTMLPlugin) Name() string        { return "dummy_html" }
func (HTMLPlugin) Description() string { return "HTML display plugin test" }
func (HTMLPlugin) Format() string      { return "html" }

// Render simply wraps Title + Excerpt inside a minimal HTML layout.
func (HTMLPlugin) Render(ctx context.Context, doc *plugins.Document) ([]byte, error) {
	html := "<html><body><h1>" + doc.Title + "</h1><p>" + doc.Excerpt + "</p></body></html>"
	return []byte(html), nil
}

//
// ───────────────────────────────────────────────
//                     MAIN
// ───────────────────────────────────────────────
//

func main() {
	cli, err := aether.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Register custom HTML DisplayPlugin
	if err := cli.RegisterDisplayPlugin(HTMLPlugin{}); err != nil {
		log.Fatal(err)
	}

	// Perform a real lookup (Wikipedia fallback)
	sr, err := cli.Search(context.Background(), "Finland")
	if err != nil {
		log.Fatal(err)
	}

	// Render using the custom HTML plugin
	fmt.Println("\n=== Render with custom HTML plugin ===")
	out, err := cli.RenderSearchResult(context.Background(), "html", sr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	// Render using an unsupported format → expect strict-mode error
	fmt.Println("\n=== Render with unsupported format ===")
	_, err = cli.Render(context.Background(), "pdf", cli.NormalizeSearchResult(sr))
	if err != nil {
		fmt.Println("Correct error:", err)
	} else {
		fmt.Println("ERROR: strict mode failed — expected a plugin-not-found error")
	}
}
