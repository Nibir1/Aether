package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

// -----------------------------
// Dummy HTML Display Plugin
// -----------------------------
type HTMLPlugin struct{}

func (HTMLPlugin) Name() string        { return "dummy_html" }
func (HTMLPlugin) Description() string { return "HTML display plugin test" }
func (HTMLPlugin) Format() string      { return "html" }

func (HTMLPlugin) Render(ctx context.Context, doc *plugins.Document) ([]byte, error) {
	html := "<html><body><h1>" + doc.Title + "</h1><p>" + doc.Excerpt + "</p></body></html>"
	return []byte(html), nil
}

func main() {
	cli, err := aether.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Register display plugin
	err = cli.RegisterDisplayPlugin(HTMLPlugin{})
	if err != nil {
		log.Fatal(err)
	}

	sr, err := cli.Search(context.Background(), "Finland")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Render with custom HTML plugin ===")
	out, err := cli.RenderSearchResult(context.Background(), "html", sr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))

	// Confirm strict mode error
	fmt.Println("\n=== Render with unsupported format ===")
	_, err = cli.Render(context.Background(), "pdf", cli.NormalizeSearchResult(sr))
	if err != nil {
		fmt.Println("Correct error:", err)
	}
}
