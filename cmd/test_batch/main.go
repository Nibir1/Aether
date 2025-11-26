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

	fmt.Println("=== Test 7: Batch Fetch ===")

	urls := []string{
		"https://example.com/",
		"https://www.wikipedia.org/",
		"https://news.ycombinator.com/",
	}

	res, err := cli.Batch(ctx, urls, aether.BatchOptions{
		Concurrency: 2,
	})
	if err != nil {
		log.Fatalf("Batch error: %v", err)
	}

	for i, r := range res.Results {
		fmt.Printf("Result %d:\n", i+1)
		fmt.Printf("  URL:        %s\n", r.URL)
		if r.Err != nil {
			fmt.Printf("  ERROR:      %v\n\n", r.Err)
			continue
		}
		fmt.Printf("  StatusCode: %d\n", r.StatusCode)
		fmt.Printf("  Body len:   %d\n\n", len(r.Body))
	}
}
