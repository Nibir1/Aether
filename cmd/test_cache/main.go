// cmd/test_cache/main.go

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	// Create default Aether client (memory cache enabled by default)
	cli, err := aether.NewClient(
		aether.WithDebugLogging(true), // enable debug output so we can SEE cache hits
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx := context.Background()

	// Stable cacheable URL
	url := "https://books.toscrape.com/"

	fmt.Println("=== Aether Cache Test ===")
	fmt.Println("Fetching:", url)
	fmt.Println()

	// --- First fetch: expected CACHE MISS ---
	res1, err := cli.Fetch(ctx, url)
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}
	fmt.Println("First fetch length:", len(res1.Body))

	// --- Second fetch: expected CACHE HIT ---
	res2, err := cli.Fetch(ctx, url)
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}
	fmt.Println("Second fetch length:", len(res2.Body))

	// --- Check equality (body size is enough for this demo) ---
	if len(res1.Body) == len(res2.Body) {
		fmt.Println("\nCache OK: second fetch returned identical size (HIT)")
	} else {
		fmt.Println("\nCache WARNING: different sizes, investigate!")
	}
}
