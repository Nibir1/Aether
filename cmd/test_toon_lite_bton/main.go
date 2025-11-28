// cmd/test_toon_lite_bton/main.go
//
// Test: TOON Lite JSON + Pretty JSON + BT0N binary serialization.
//
// This test performs a search, normalizes the SearchResult, and then
// serializes it using TOON Lite (compact + pretty) and BT0N formats.
// A round-trip decode of BT0N is also performed to validate integrity.

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
	fmt.Println("====================================================")
	fmt.Println("TEST: TOON Lite + BTON — NORMAL MODE (robots ON)")
	fmt.Println("====================================================")

	runTOONLiteBTONTest(false)

	fmt.Println("\n====================================================")
	fmt.Println("TEST: TOON Lite + BTON — ROBOTS OVERRIDE ENABLED")
	fmt.Println("====================================================")

	runTOONLiteBTONTest(true)
}

func runTOONLiteBTONTest(override bool) {
	// -----------------------------
	// Create Aether client
	// -----------------------------
	var cli *aether.Client
	var err error

	if override {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
			aether.WithRobotsOverride("hnrss.org", "news.ycombinator.com"),
		)
	} else {
		cli, err = aether.NewClient(
			aether.WithDebugLogging(true),
		)
	}

	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	// -----------------------------
	// 1) Perform Search
	// -----------------------------
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// -----------------------------
	// 2) Normalize SearchResult
	// -----------------------------
	norm := cli.NormalizeSearchResult(sr)
	if norm == nil {
		log.Fatalf("NormalizeSearchResult returned nil")
	}

	// -------------------------------------------------------------------------
	// 3) TOON Lite (compact)
	// -------------------------------------------------------------------------
	fmt.Println("-> MarshalTOONLite (compact):")
	tlite, err := cli.MarshalTOONLite(sr)
	if err != nil {
		log.Fatalf("MarshalTOONLite error: %v", err)
	}
	fmt.Println(string(tlite))

	// -------------------------------------------------------------------------
	// 4) TOON Lite Pretty
	// -------------------------------------------------------------------------
	fmt.Println("\n-> MarshalTOONLitePretty (pretty JSON):")
	tlitePretty, err := cli.MarshalTOONLitePretty(sr)
	if err != nil {
		log.Fatalf("MarshalTOONLitePretty error: %v", err)
	}
	fmt.Println(string(tlitePretty))

	// -------------------------------------------------------------------------
	// 5) BT0N encoding
	// -------------------------------------------------------------------------
	fmt.Println("\n-> MarshalBTON (binary TOON):")
	btonBytes, err := cli.MarshalBTON(sr)
	if err != nil {
		log.Fatalf("MarshalBTON error: %v", err)
	}
	fmt.Printf("BTON bytes (%d bytes):\n%s\n",
		len(btonBytes),
		hex.Dump(btonBytes))

	// Optionally write BT0N to disk
	_ = os.WriteFile("test_output.bton", btonBytes, 0644)
	fmt.Println("Saved to test_output.bton")

	// -------------------------------------------------------------------------
	// 6) BT0N decode round-trip test
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
	fmt.Println()
}
