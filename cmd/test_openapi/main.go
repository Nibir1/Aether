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
	//  A) TEST WITH NORMAL CLIENT — robots.txt respected (default)
	// ================================================================
	fmt.Println("\n====================================================")
	fmt.Println("A) OPENAPI TESTS — ROBOTS.TXT MODE (DEFAULT)")
	fmt.Println("====================================================")

	cliNormal, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create normal Aether client: %v", err)
	}

	runOpenAPITests("NORMAL MODE (robots ON)", cliNormal)

	// ================================================================
	//  B) TEST WITH ROBOTS OVERRIDE CLIENT — Hacker News bypassed
	// ================================================================
	//
	// HN denies API crawling via robots.txt. We explicitly override it
	// ONLY for *.ycombinator.com — everything else stays safe.
	//
	// This mirrors Option A (legal opt-in: user takes responsibility).
	// ================================================================

	fmt.Println("\n====================================================")
	fmt.Println("B) OPENAPI TESTS — ROBOTS OVERRIDE FOR HN")
	fmt.Println("====================================================")

	cliOverride, err := aether.NewClient(
		aether.WithDebugLogging(true),
		aether.WithRobotsOverride("news.ycombinator.com", "hacker-news.firebaseio.com"),
	)
	if err != nil {
		log.Fatalf("failed to create override Aether client: %v", err)
	}

	runOpenAPITests("ROBOTS OVERRIDE MODE (HN allowed)", cliOverride)
}

// ------------------------------------------------------------
// Shared Test Runner
// ------------------------------------------------------------
func runOpenAPITests(label string, cli *aether.Client) {
	fmt.Println("\n>>> Running test suite for:", label)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
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
	// 2) Hacker News — THIS IS WHERE ROBOTS OVERRIDE MATTERS
	// --------------------------------------------------------
	fmt.Println("-> HackerNewsTopStories(limit=3)")
	hn, err := cli.HackerNewsTopStories(ctx, 3)
	if err != nil {
		log.Printf("HackerNewsTopStories error: %v", err)
	} else if hn == nil {
		fmt.Println("  HackerNewsTopStories returned nil")
	} else {
		for i, s := range hn {
			fmt.Printf("  Story %d: [%d] %s (score=%d, by=%s)\n",
				i+1, s.ID, s.Title, s.Score, s.Author)
			fmt.Printf("    URL: %s\n", s.URL)
		}
		fmt.Println()
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
