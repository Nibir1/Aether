// cmd/test_search_display/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Nibir1/Aether/aether"
)

func main() {

	// =====================================================================
	// A) NORMAL MODE — robots.txt respected
	// =====================================================================
	fmt.Println("====================================================")
	fmt.Println("A) SEARCH + DISPLAY TEST — ROBOTS.TXT MODE (DEFAULT)")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: NORMAL MODE (robots ON)")
	fmt.Println()
	runSearchDisplayTest(false)

	// =====================================================================
	// B) OVERRIDE MODE — allow Hacker News (for example)
	// =====================================================================
	fmt.Println("====================================================")
	fmt.Println("B) SEARCH + DISPLAY TEST — ROBOTS OVERRIDE ENABLED")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: ROBOTS OVERRIDE MODE")
	fmt.Println()
	runSearchDisplayTest(true)
}

func runSearchDisplayTest(override bool) {

	var cli *aether.Client
	var err error

	if override {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
			// Example: allow known sources for override
			aether.WithRobotsOverride("news.ycombinator.com", "hnrss.org"),
		)
	} else {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
		)
	}

	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := "Finland"
	fmt.Println("Query:", query)
	fmt.Println()

	// 1) Search
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Printf("Search error: %v\n", err)
		return
	}

	fmt.Println("Search Plan:")
	fmt.Printf("  Intent: %s\n", sr.Plan.Intent)
	fmt.Printf("  Source: %s\n", sr.Plan.Source)
	if sr.PrimaryDocument != nil {
		fmt.Printf("  URL:    %s\n", sr.PrimaryDocument.URL)
	}
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
		log.Printf("Render(markdown) error: %v\n", err)
	} else {
		fmt.Println("Rendered Markdown (first 400 chars):")
		if len(md) > 400 {
			fmt.Println(string(md[:400]))
		} else {
			fmt.Println(string(md))
		}
		fmt.Println()
	}

	preview, err := cli.Render(context.Background(), "preview", norm)
	if err != nil {
		log.Printf("Render(preview) error: %v\n", err)
	} else {
		fmt.Println("Rendered Preview:")
		fmt.Println(string(preview))
		fmt.Println()
	}

	// 4) JSON normalization output
	jsonBytes, err := cli.MarshalSearchResultJSON(sr)
	if err != nil {
		log.Printf("MarshalSearchResultJSON error: %v\n", err)
	} else {
		fmt.Println("JSON (first 400 chars):")
		if len(jsonBytes) > 400 {
			fmt.Println(string(jsonBytes[:400]))
		} else {
			fmt.Println(string(jsonBytes))
		}
		fmt.Println()
	}

	// 5) TOON JSON
	toonBytes, err := cli.MarshalSearchResultTOON(sr)
	if err != nil {
		log.Printf("MarshalSearchResultTOON error: %v\n", err)
	} else {
		fmt.Println("TOON JSON (first 400 chars):")
		if len(toonBytes) > 400 {
			fmt.Println(string(toonBytes[:400]))
		} else {
			fmt.Println(string(toonBytes))
		}
		fmt.Println()
	}
}
