package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	// --- Create default Aether client (no custom options needed) ---
	cli, err := aether.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// A URL that is guaranteed to be cacheable (public + stable)
	url := "https://example.com"

	fmt.Println("=== Aether Cache Test ===")
	fmt.Println("Fetching:", url)
	fmt.Println()

	// --- First fetch: CACHE MISS ---
	body1, _, err := cli.FetchRaw(ctx, url)
	if err != nil {
		log.Fatalf("FetchRaw error: %v", err)
	}
	fmt.Println("First fetch length:", len(body1))

	// --- Second fetch: CACHE HIT ---
	body2, _, err := cli.FetchRaw(ctx, url)
	if err != nil {
		log.Fatalf("FetchRaw error: %v", err)
	}
	fmt.Println("Second fetch length:", len(body2))

	// --- Check equality ---
	if len(body1) == len(body2) {
		fmt.Println("\nCache OK: second fetch returned identical size (HIT)")
	} else {
		fmt.Println("\nCache WARNING: different sizes, investigate!")
	}
}
