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

	fmt.Println("=== Test 6: TOON Streaming ===")

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	norm := cli.NormalizeSearchResult(sr)

	fmt.Println("-> StreamTOON (NormalizedDocument) to stdout:")
	if err := cli.StreamTOON(context.Background(), os.Stdout, norm); err != nil {
		log.Fatalf("StreamTOON error: %v", err)
	}

	fmt.Println("\n-> StreamSearchResultTOON to stdout:")
	if err := cli.StreamSearchResultTOON(context.Background(), os.Stdout, sr); err != nil {
		log.Fatalf("StreamSearchResultTOON error: %v", err)
	}
}
