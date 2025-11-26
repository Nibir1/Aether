package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
	"github.com/Nibir1/Aether/plugins/examples/hn_plugin"
)

// simpleTransformPlugin uppercases the document title and adds metadata.
type simpleTransformPlugin struct{}

func (p *simpleTransformPlugin) Name() string { return "simple_transform_upper_title" }
func (p *simpleTransformPlugin) Description() string {
	return "Uppercases document title and tags metadata."
}

func (p *simpleTransformPlugin) Apply(ctx context.Context, doc *plugins.Document) (*plugins.Document, error) {
	if doc == nil {
		return nil, nil
	}
	out := *doc // shallow copy
	out.Title = strings.ToUpper(doc.Title)
	if out.Metadata == nil {
		out.Metadata = map[string]string{}
	}
	out.Metadata["transform.upper_title"] = "true"
	return &out, nil
}

// upperDisplayPlugin renders the title and excerpt in uppercase plain text.
type upperDisplayPlugin struct{}

func (p *upperDisplayPlugin) Name() string { return "upper_display" }
func (p *upperDisplayPlugin) Description() string {
	return "Renders document title + excerpt in uppercase text."
}
func (p *upperDisplayPlugin) Format() string { return "upper" }

func (p *upperDisplayPlugin) Render(ctx context.Context, doc *plugins.Document) ([]byte, error) {
	if doc == nil {
		return []byte("EMPTY DOCUMENT"), nil
	}
	title := strings.ToUpper(doc.Title)
	excerpt := strings.ToUpper(doc.Excerpt)
	out := fmt.Sprintf("TITLE: %s\nEXCERPT: %s\n", title, excerpt)
	return []byte(out), nil
}

func main() {
	cli, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("=== Test 9: Plugins (Source + Transform + Display) ===")

	// 1) Register Source plugin (HackerNews)
	hn := hn_plugin.New(cli, 5)
	if err := cli.RegisterSourcePlugin(hn); err != nil {
		log.Fatalf("RegisterSourcePlugin error: %v", err)
	}
	fmt.Println("Registered SourcePlugin: hackernews")

	// 2) Register Transform plugin
	tp := &simpleTransformPlugin{}
	if err := cli.RegisterTransformPlugin(tp); err != nil {
		log.Fatalf("RegisterTransformPlugin error: %v", err)
	}
	fmt.Println("Registered TransformPlugin:", tp.Name())

	// 3) Register Display plugin
	dp := &upperDisplayPlugin{}
	if err := cli.RegisterDisplayPlugin(dp); err != nil {
		log.Fatalf("RegisterDisplayPlugin error: %v", err)
	}
	fmt.Println("Registered DisplayPlugin:", dp.Name(), "format:", dp.Format())
	fmt.Println()

	// 4) Run Search that should hit HackerNews via SourcePlugin
	query := "latest tech news"
	fmt.Println("Search query:", query)
	fmt.Println()

	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	fmt.Println("Search Plan:")
	fmt.Printf("  Intent: %s\n", sr.Plan.Intent)
	fmt.Printf("  Source: %s\n", sr.Plan.Source)
	fmt.Println()

	// 5) Normalize (this will apply TransformPlugins)
	norm := cli.NormalizeSearchResult(sr)
	fmt.Println("Normalized Document after TransformPlugins:")
	fmt.Printf("  Title: %s\n", norm.Title)
	fmt.Printf("  Excerpt: %s\n", norm.Excerpt)
	fmt.Printf("  Metadata[\"transform.upper_title\"]: %s\n", norm.Metadata["transform.upper_title"])
	fmt.Println()

	// 6) Render using DisplayPlugin ("upper" format)
	fmt.Println("Render with DisplayPlugin format=\"upper\":")
	out, err := cli.Render(context.Background(), "upper", norm)
	if err != nil {
		log.Fatalf("Render(upper) error: %v", err)
	}
	fmt.Println(string(out))
}
