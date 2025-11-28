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

//
// ────────────────────────────────────────────────
// Simple Transform Plugin (Uppercase Title)
// ────────────────────────────────────────────────
//

type simpleTransformPlugin struct{}

func (p *simpleTransformPlugin) Name() string { return "simple_transform_upper_title" }
func (p *simpleTransformPlugin) Description() string {
	return "Uppercases document title and tags metadata."
}

func (p *simpleTransformPlugin) Apply(ctx context.Context, doc *plugins.Document) (*plugins.Document, error) {
	if doc == nil {
		return nil, nil
	}
	out := *doc
	out.Title = strings.ToUpper(doc.Title)
	if out.Metadata == nil {
		out.Metadata = map[string]string{}
	}
	out.Metadata["transform.upper_title"] = "true"
	return &out, nil
}

//
// ────────────────────────────────────────────────
// Simple Display Plugin (Uppercase Output)
// ────────────────────────────────────────────────
//

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

//
// ────────────────────────────────────────────────
// Main: Show Both Modes (Robots ON vs Robots Override)
// ────────────────────────────────────────────────
//

func main() {

	// =====================================================
	// A) NORMAL MODE (robots enforced)
	// =====================================================
	fmt.Println("=====================================================")
	fmt.Println("A) PLUGIN TEST — ROBOTS.TXT MODE (DEFAULT)")
	fmt.Println("=====================================================")
	runPluginTest(false)

	// Small gap
	fmt.Println("\n\n=====================================================")
	fmt.Println("B) PLUGIN TEST — ROBOTS OVERRIDE ENABLED FOR HN")
	fmt.Println("=====================================================")
	runPluginTest(true)
}

//
// runPluginTest(overridden bool)
// If overridden=false → normal robots.txt enforcement
// If overridden=true  → allow hacker-news.firebaseio.com explicitly
//

func runPluginTest(override bool) {

	options := []aether.Option{
		aether.WithDebugLogging(true),
	}

	if override {
		options = append(options,
			aether.WithRobotsOverride("hacker-news.firebaseio.com"),
		)
	}

	cli, err := aether.NewClient(options...)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 1) Register HackerNews Source Plugin
	hn := hn_plugin.New(cli, 5)
	if err := cli.RegisterSourcePlugin(hn); err != nil {
		log.Fatalf("RegisterSourcePlugin error: %v", err)
	}

	fmt.Printf("\n>>> Running test suite for: %s\n",
		func() string {
			if override {
				return "ROBOTS OVERRIDE MODE"
			}
			return "NORMAL MODE (robots ON)"
		}())

	// 2) Register Transform Plugin
	tp := &simpleTransformPlugin{}
	if err := cli.RegisterTransformPlugin(tp); err != nil {
		log.Fatalf("RegisterTransformPlugin error: %v", err)
	}

	// 3) Register Display Plugin
	dp := &upperDisplayPlugin{}
	if err := cli.RegisterDisplayPlugin(dp); err != nil {
		log.Fatalf("RegisterDisplayPlugin error: %v", err)
	}

	// 4) Run Search
	query := "latest tech news"
	fmt.Println("\nSearch query:", query)

	sr, err := cli.Search(ctx, query)
	if err != nil {
		fmt.Println("Search error:", err)
		return
	}

	fmt.Println("Search Plan:")
	fmt.Printf("  Intent: %s\n", sr.Plan.Intent)
	fmt.Printf("  Source: %s\n", sr.Plan.Source)

	// 5) Normalize (applies transform plugins)
	norm := cli.NormalizeSearchResult(sr)

	fmt.Println("\nNormalized Document after TransformPlugins:")
	fmt.Printf("  Title: %s\n", norm.Title)
	fmt.Printf("  Excerpt: %s\n", norm.Excerpt)
	fmt.Printf("  Metadata[\"transform.upper_title\"]: %s\n",
		norm.Metadata["transform.upper_title"])

	// 6) Render via Display Plugin
	fmt.Println("\nRender with DisplayPlugin format=\"upper\":")
	out, err := cli.Render(context.Background(), "upper", norm)
	if err != nil {
		fmt.Println("Render error:", err)
		return
	}

	fmt.Println(string(out))
}
