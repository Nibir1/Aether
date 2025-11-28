// cmd/test_rss/main.go

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
	// A) NORMAL MODE (robots.txt ON, no overrides)
	// =====================================================================

	fmt.Println("====================================================")
	fmt.Println("A) RSS TEST — ROBOTS.TXT MODE (DEFAULT)")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: NORMAL MODE (robots ON)")
	fmt.Println()

	runRSSTest(false) // robots override disabled

	// =====================================================================
	// B) OVERRIDE MODE (allow HN feed)
	// =====================================================================

	fmt.Println("====================================================")
	fmt.Println("B) RSS TEST — ROBOTS OVERRIDE ENABLED FOR HN")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: ROBOTS OVERRIDE MODE")
	fmt.Println()

	runRSSTest(true) // override enabled
}

func runRSSTest(override bool) {

	// -----------------------------
	// Create client with optional overrides
	// -----------------------------
	var cli *aether.Client
	var err error

	if override {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
			aether.WithRobotsOverride(
				"hnrss.org",
				"news.ycombinator.com",
			),
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
	// Context with timeout
	// -----------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// -----------------------------
	// Target feed
	// -----------------------------
	feedURL := "https://hnrss.org/frontpage"

	fmt.Println("=== Test: RSS Fetch + Parse ===")
	fmt.Println("Feed URL:", feedURL)
	fmt.Println()

	// -----------------------------
	// Fetch + Parse RSS
	// -----------------------------
	feed, err := cli.FetchRSS(ctx, feedURL)
	if err != nil {
		fmt.Printf("FetchRSS error: %v\n\n", err)
		return
	}

	fmt.Println("Feed:")
	fmt.Printf("  Title:       %s\n", feed.Title)
	fmt.Printf("  Description: %s\n", feed.Description)
	fmt.Printf("  Link:        %s\n", feed.Link)
	fmt.Printf("  Items:       %d\n", len(feed.Items))
	fmt.Println()

	for i, item := range feed.Items {
		if i >= 5 {
			break
		}
		fmt.Printf("Item %d:\n", i+1)
		fmt.Printf("  Title: %s\n", item.Title)
		fmt.Printf("  Link:  %s\n", item.Link)
		fmt.Printf("  GUID:  %s\n", item.GUID)
		fmt.Println()
	}
}
