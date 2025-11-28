// cmd/test_jsonl/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	cli, err := aether.NewClient(
		aether.WithDebugLogging(false),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println("=== Test: JSONL Streaming ===")

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	// Perform search
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// Normalize the result
	norm := cli.NormalizeSearchResult(sr)
	if norm == nil {
		log.Fatalf("NormalizeSearchResult returned nil")
	}

	// ------------------------------------------------------------
	// 1. StreamNormalizedJSONL (normalized Document → JSONL)
	// ------------------------------------------------------------
	fmt.Println("-> StreamNormalizedJSONL (NormalizedDocument) to stdout:")
	err = cli.StreamNormalizedJSONL(context.Background(), os.Stdout, norm)
	if err != nil {
		log.Fatalf("StreamNormalizedJSONL error: %v", err)
	}

	// ------------------------------------------------------------
	// 2. StreamSearchResultJSONL
	// ------------------------------------------------------------
	fmt.Println("\n-> StreamSearchResultJSONL (SearchResult → JSONL):")
	err = cli.StreamSearchResultJSONL(context.Background(), os.Stdout, sr)
	if err != nil {
		log.Fatalf("StreamSearchResultJSONL error: %v", err)
	}

	// ------------------------------------------------------------
	// 3. Stream feed items (if present)
	// ------------------------------------------------------------
	if sr.Feed != nil && len(sr.Feed.Items) > 0 {
		fmt.Println("\n-> StreamFeedJSONL (Raw feed items):")
		err = cli.StreamFeedJSONL(context.Background(), os.Stdout, sr.Feed)
		if err != nil {
			log.Fatalf("StreamFeedJSONL error: %v", err)
		}
	} else {
		fmt.Println("\n(No feed items detected — skipping StreamFeedJSONL)")
	}
}
