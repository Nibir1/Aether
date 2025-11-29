// cmd/test_openapi/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Nibir1/Aether/aether"
)

func main() {

	// ================================================================
	//  A) NORMAL CLIENT — robots.txt respected
	// ================================================================
	fmt.Println("\n====================================================")
	fmt.Println("OPENAPI TESTS — NORMAL CLIENT (robots.txt respected)")
	fmt.Println("====================================================")

	cliNormal, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create normal Aether client: %v", err)
	}

	runOpenAPITests("NORMAL CLIENT", cliNormal)

	// ================================================================
	//  B) OVERRIDE CLIENT — Hacker News allowed
	// ================================================================
	fmt.Println("\n====================================================")
	fmt.Println("OPENAPI TESTS — ROBOTS OVERRIDE CLIENT (HN allowed)")
	fmt.Println("====================================================")

	cliOverride, err := aether.NewClient(
		aether.WithDebugLogging(true),
		aether.WithRobotsOverride("hacker-news.firebaseio.com"),
	)
	if err != nil {
		log.Fatalf("failed to create override Aether client: %v", err)
	}

	runOpenAPITests("OVERRIDE CLIENT", cliOverride)
}

// ------------------------------------------------------------
// Shared Test Runner
// ------------------------------------------------------------
func runOpenAPITests(label string, cli *aether.Client) {
	fmt.Println("\n>>> Running OpenAPI test suite for:", label)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --------------------------------------------------------
	// 1) Wikipedia Summary
	// --------------------------------------------------------
	fmt.Println("\n-> WikipediaSummary(\"Finland\")")
	wiki, err := cli.WikipediaSummary(ctx, "Finland")
	if err != nil {
		log.Printf("WikipediaSummary error: %v", err)
	} else if wiki == nil {
		fmt.Println("  WikipediaSummary returned nil")
	} else {
		fmt.Printf("  Title: %s\n", wiki.Title)
		fmt.Printf("  Description: %s\n", wiki.Description)
		if len(wiki.Extract) > 200 {
			fmt.Printf("  Extract (first 200 chars): %s\n", wiki.Extract[:200])
		} else {
			fmt.Printf("  Extract: %s\n", wiki.Extract)
		}
		fmt.Printf("  URL: %s\n\n", wiki.URL)
	}

	// --------------------------------------------------------
	// 2) Hacker News Top Stories — Normalized Documents
	// --------------------------------------------------------
	fmt.Println("-> HackerNewsTopStoriesDocuments(limit=3)")
	hnDocs, err := cli.HackerNewsTopStoriesDocuments(ctx, 3)
	if err != nil {
		log.Printf("HackerNewsTopStoriesDocuments error: %v", err)
	} else if hnDocs == nil {
		fmt.Println("  HackerNewsTopStoriesDocuments returned nil")
	} else {
		for i, doc := range hnDocs {
			fmt.Printf("  Story %d: %s\n", i+1, doc.Title)
			fmt.Printf("    Excerpt: %s\n", doc.Excerpt)
			fmt.Printf("    URL: %s\n", doc.SourceURL)
			fmt.Printf("    Metadata: %+v\n\n", doc.Metadata)
		}
	}

	// --------------------------------------------------------
	// 3) GitHub README (golang/go)
	// --------------------------------------------------------
	fmt.Println("-> GitHubReadme(\"golang\", \"go\", \"master\")")
	readme, err := cli.GitHubReadme(ctx, "golang", "go", "master")
	if err != nil {
		log.Printf("GitHubReadme error: %v", err)
	} else if readme == nil {
		fmt.Println("  GitHubReadme returned nil")
	} else {
		fmt.Printf("  Repo: %s/%s\n", readme.Owner, readme.Repo)
		fmt.Printf("  URL:  %s\n", readme.URL)
		if len(readme.Content) > 200 {
			fmt.Printf("  Content (first 200 chars): %s\n\n", readme.Content[:200])
		} else {
			fmt.Printf("  Content: %s\n\n", readme.Content)
		}
	}

	// --------------------------------------------------------
	// 4) WeatherAt — using MET Norway OpenAPI
	// --------------------------------------------------------
	fmt.Println("-> WeatherAt(60.1699, 24.9384, hours=6)")
	weather, err := cli.WeatherAt(ctx, 60.1699, 24.9384, 6)
	if err != nil {
		log.Printf("WeatherAt error: %v", err)
	} else if weather == nil {
		fmt.Println("  WeatherAt returned nil")
	} else {
		for i, w := range weather {
			if i >= 3 {
				break
			}
			fmt.Printf("  Entry %d: temp=%.1f°C wind=%.1fm/s summary=%s\n",
				i+1, w.Temperature, w.WindSpeed, w.Summary)
		}
	}
}
