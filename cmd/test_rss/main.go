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

	// A well-known example RSS feed
	feedURL := "https://hnrss.org/frontpage"

	fmt.Println("=== Test 2: RSS Fetch + Parse ===")
	fmt.Println("Feed URL:", feedURL)
	fmt.Println()

	feed, err := cli.FetchRSS(ctx, feedURL)
	if err != nil {
		log.Fatalf("FetchRSS error: %v", err)
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
