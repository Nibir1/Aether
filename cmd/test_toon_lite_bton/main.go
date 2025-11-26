package main

import (
	"context"
	"encoding/hex"
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

	fmt.Println("=== Test 10: TOON Lite + BTON ===")

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	// Perform search
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// Normalize first (good practice)
	norm := cli.NormalizeSearchResult(sr)
	if norm == nil {
		log.Fatalf("NormalizeSearchResult returned nil")
	}

	// -------------------------------------------------------------------------
	// 1. TOON Lite (compact)
	// -------------------------------------------------------------------------
	fmt.Println("\n-> MarshalTOONLite (compact):")
	tlite, err := cli.MarshalTOONLite(sr)
	if err != nil {
		log.Fatalf("MarshalTOONLite error: %v", err)
	}
	fmt.Println(string(tlite))

	// -------------------------------------------------------------------------
	// 2. TOON Lite Pretty
	// -------------------------------------------------------------------------
	fmt.Println("\n-> MarshalTOONLitePretty (pretty JSON):")
	tlitePretty, err := cli.MarshalTOONLitePretty(sr)
	if err != nil {
		log.Fatalf("MarshalTOONLitePretty error: %v", err)
	}
	fmt.Println(string(tlitePretty))

	// -------------------------------------------------------------------------
	// 3. BTON encoding
	// -------------------------------------------------------------------------
	fmt.Println("\n-> MarshalBTON (binary TOON):")
	btonBytes, err := cli.MarshalBTON(sr)
	if err != nil {
		log.Fatalf("MarshalBTON error: %v", err)
	}
	fmt.Printf("BTON bytes (%d bytes):\n%s\n",
		len(btonBytes),
		hex.Dump(btonBytes))

	// Optionally write BTON to disk
	_ = os.WriteFile("test_output.bton", btonBytes, 0644)
	fmt.Println("Saved to test_output.bton")

	// -------------------------------------------------------------------------
	// 4. BTON decode round-trip test
	// -------------------------------------------------------------------------
	fmt.Println("\n-> UnmarshalBTON (decode round-trip):")
	decoded, err := cli.UnmarshalBTON(btonBytes)
	if err != nil {
		log.Fatalf("UnmarshalBTON error: %v", err)
	}

	fmt.Println("Decoded TOON Document:")
	fmt.Printf("Kind: %s\n", decoded.Kind)
	fmt.Printf("Title: %s\n", decoded.Title)
	fmt.Printf("Excerpt: %s\n", decoded.Excerpt)
	fmt.Printf("Tokens: %d tokens\n", len(decoded.Tokens))
}
