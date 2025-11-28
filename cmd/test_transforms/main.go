// cmd/test_transforms/main.go
//
// Test case for Aether Transform Plugins
// Demonstrates applying a dummy transform plugin that uppercases the title.

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

// -----------------------------
// Dummy Transform Plugin
// -----------------------------
type UppercaseTitleTransform struct{}

func (UppercaseTitleTransform) Name() string { return "uppercase_title" }
func (UppercaseTitleTransform) Description() string {
	return "Uppercases the document title for testing purposes."
}

// Apply uppercases the document title
func (UppercaseTitleTransform) Apply(ctx context.Context, doc *plugins.Document) (*plugins.Document, error) {
	if doc == nil {
		return nil, nil
	}
	doc.Title = strings.ToUpper(doc.Title)
	if doc.Metadata == nil {
		doc.Metadata = map[string]string{}
	}
	doc.Metadata["transform.uppercase_title"] = "true"
	return doc, nil
}

// -----------------------------
// Test Runner
// -----------------------------
func main() {
	fmt.Println("=== Test: Transform Plugins (Normal + Robots Override Modes) ===")

	// Run both normal mode and robots override mode
	runTransformTest(false) // normal mode
	fmt.Println()
	runTransformTest(true) // override mode
}

// runTransformTest executes the test with optional robots override
func runTransformTest(override bool) {
	mode := "NORMAL MODE (robots ON)"
	if override {
		mode = "ROBOTS OVERRIDE ENABLED"
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("Test Mode:", mode)
	fmt.Println("--------------------------------------------------")

	// -----------------------------
	// Create Aether client
	// -----------------------------
	var cli *aether.Client
	var err error

	if override {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
			aether.WithRobotsOverride("hnrss.org", "news.ycombinator.com"),
		)
	} else {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
		)
	}
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	// -----------------------------
	// Register Transform Plugin
	// -----------------------------
	err = cli.RegisterTransformPlugin(UppercaseTitleTransform{})
	if err != nil {
		log.Fatalf("RegisterTransformPlugin error: %v", err)
	}

	// -----------------------------
	// Execute Search
	// -----------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// -----------------------------
	// Normalize Search Result (applies transform plugins)
	// -----------------------------
	normalized := cli.NormalizeSearchResult(sr)

	fmt.Println("Normalized Document after TransformPlugins:")
	fmt.Printf("  Title:   %s\n", normalized.Title)
	fmt.Printf("  Excerpt: %s\n", normalized.Excerpt)
	if val, ok := normalized.Metadata["transform.uppercase_title"]; ok {
		fmt.Println("  Metadata[transform.uppercase_title]:", val)
	}
}
