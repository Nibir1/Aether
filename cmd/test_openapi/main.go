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

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	fmt.Println("=== Test 3: OpenAPI Integrations ===")

	// 1) WikipediaSummary
	fmt.Println("-> WikipediaSummary(\"Finland\")")
	wiki, err := cli.WikipediaSummary(ctx, "Finland")
	if err != nil {
		log.Printf("WikipediaSummary error: %v", err)
	} else if wiki != nil {
		fmt.Printf("  Title: %s\n", wiki.Title)
		fmt.Printf("  Description: %s\n", wiki.Description)
		if len(wiki.Extract) > 200 {
			fmt.Printf("  Extract (first 200 chars): %s\n", wiki.Extract[:200])
		} else {
			fmt.Printf("  Extract: %s\n", wiki.Extract)
		}
		fmt.Printf("  URL: %s\n\n", wiki.URL)
	}

	// 2) HackerNewsTopStories
	fmt.Println("-> HackerNewsTopStories(limit=3)")
	hn, err := cli.HackerNewsTopStories(ctx, 3)
	if err != nil {
		log.Printf("HackerNewsTopStories error: %v", err)
	} else {
		for i, s := range hn {
			fmt.Printf("  Story %d: [%d] %s (score=%d, by=%s)\n", i+1, s.ID, s.Title, s.Score, s.Author)
			fmt.Printf("    URL: %s\n", s.URL)
		}
		fmt.Println()
	}

	// 3) GitHubReadme
	fmt.Println("-> GitHubReadme(\"golang\", \"go\", \"master\")")
	readme, err := cli.GitHubReadme(ctx, "golang", "go", "master")
	if err != nil {
		log.Printf("GitHubReadme error: %v", err)
	} else if readme != nil {
		fmt.Printf("  Repo: %s/%s\n", readme.Owner, readme.Repo)
		fmt.Printf("  URL:  %s\n", readme.URL)
		if len(readme.Content) > 200 {
			fmt.Printf("  Content (first 200 chars): %s\n\n", readme.Content[:200])
		} else {
			fmt.Printf("  Content: %s\n\n", readme.Content)
		}
	}

	// 4) WeatherAt (Helsinki approx: 60.1699 N, 24.9384 E)
	fmt.Println("-> WeatherAt(60.1699, 24.9384, hours=6)")
	weather, err := cli.WeatherAt(ctx, 60.1699, 24.9384, 6)
	if err != nil {
		log.Printf("WeatherAt error: %v", err)
	} else {
		for i, w := range weather {
			if i >= 3 {
				break
			}
			fmt.Printf("  Entry %d: temp=%.1f°C wind=%.1fm/s summary=%s\n", i+1, w.Temperature, w.WindSpeed, w.Summary)
		}
	}
}
