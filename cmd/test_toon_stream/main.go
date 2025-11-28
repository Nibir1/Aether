// cmd/test_toon_stream/main.go

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

	// =====================================================================
	// A) NORMAL MODE (robots.txt ON, no overrides)
	// =====================================================================
	fmt.Println("====================================================")
	fmt.Println("A) TOON STREAM TEST — ROBOTS.TXT MODE (DEFAULT)")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: NORMAL MODE (robots ON)")
	fmt.Println()

	runTOONStreamTest(false) // robots override disabled

	// =====================================================================
	// B) OVERRIDE MODE (allow HN feed)
	// =====================================================================
	fmt.Println("====================================================")
	fmt.Println("B) TOON STREAM TEST — ROBOTS OVERRIDE ENABLED")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println(">>> Running test suite for: ROBOTS OVERRIDE MODE")
	fmt.Println()

	runTOONStreamTest(true) // override enabled
}

func runTOONStreamTest(override bool) {

	// -----------------------------
	// Create client with optional robots override
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

	query := "Finland"
	fmt.Println("Search query:", query)
	fmt.Println()

	// -----------------------------
	// Perform search
	// -----------------------------
	sr, err := cli.Search(ctx, query)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// Normalize first (recommended)
	norm := cli.NormalizeSearchResult(sr)

	// -------------------------------------------------------------------------
	// 1. Stream normalized document (JSONL/TOON)
	// -------------------------------------------------------------------------
	fmt.Println("-> StreamTOON (NormalizedDocument) to stdout:")
	if err := cli.StreamTOON(context.Background(), os.Stdout, norm); err != nil {
		log.Fatalf("StreamTOON error: %v", err)
	}

	// -------------------------------------------------------------------------
	// 2. Stream full SearchResult (with Article/Feed if present)
	// -------------------------------------------------------------------------
	fmt.Println("\n-> StreamSearchResultTOON to stdout:")
	if err := cli.StreamSearchResultTOON(context.Background(), os.Stdout, sr); err != nil {
		log.Fatalf("StreamSearchResultTOON error: %v", err)
	}
}
