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
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("=== Test 4: Search + Normalize + Display + TOON ===")

	query := "Finland"
	fmt.Println("Query:", query)
	fmt.Println()

	// 1) Search
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	fmt.Println("Search Plan:")
	fmt.Printf("  Intent: %s\n", sr.Plan.Intent)
	fmt.Printf("  Source: %s\n", sr.Plan.Source)
	fmt.Printf("  URL:    %s\n", sr.PrimaryDocument.URL)
	fmt.Println()

	// 2) Normalize
	norm := cli.NormalizeSearchResult(sr)
	fmt.Println("Normalized Document:")
	fmt.Printf("  Kind:    %s\n", norm.Kind)
	fmt.Printf("  Title:   %s\n", norm.Title)
	fmt.Printf("  Excerpt: %s\n", norm.Excerpt)
	fmt.Println()

	// 3) Render (markdown + preview)
	md, err := cli.Render(context.Background(), "markdown", norm)
	if err != nil {
		log.Fatalf("Render(markdown) error: %v", err)
	}
	fmt.Println("Rendered Markdown (first 400 chars):")
	if len(md) > 400 {
		fmt.Println(string(md[:400]))
	} else {
		fmt.Println(string(md))
	}
	fmt.Println()

	preview, err := cli.Render(context.Background(), "preview", norm)
	if err != nil {
		log.Fatalf("Render(preview) error: %v", err)
	}
	fmt.Println("Rendered Preview:")
	fmt.Println(string(preview))
	fmt.Println()

	// 4) JSON normalization output
	jsonBytes, err := cli.MarshalSearchResultJSON(sr)
	if err != nil {
		log.Fatalf("MarshalSearchResultJSON error: %v", err)
	}
	fmt.Println("JSON (first 400 chars):")
	if len(jsonBytes) > 400 {
		fmt.Println(string(jsonBytes[:400]))
	} else {
		fmt.Println(string(jsonBytes))
	}
	fmt.Println()

	// 5) TOON JSON
	toonBytes, err := cli.MarshalSearchResultTOON(sr)
	if err != nil {
		log.Fatalf("MarshalSearchResultTOON error: %v", err)
	}
	fmt.Println("TOON JSON (first 400 chars):")
	if len(toonBytes) > 400 {
		fmt.Println(string(toonBytes[:400]))
	} else {
		fmt.Println(string(toonBytes))
	}
}
