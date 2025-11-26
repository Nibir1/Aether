package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	// Create client with debug logging enabled so we see what happens.
	cli, err := aether.NewClient(
		aether.WithDebugLogging(true),
	)
	if err != nil {
		log.Fatalf("failed to create Aether client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := "https://example.com/"

	fmt.Println("=== Test 1: Fetch + Detect + ParseHTML + ExtractArticle ===")
	fmt.Println("URL:", url)
	fmt.Println()

	// 1) Fetch
	res, err := cli.Fetch(ctx, url)
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}
	fmt.Printf("Fetch status: %d\n", res.StatusCode)
	fmt.Printf("Body length: %d bytes\n", len(res.Body))
	fmt.Println()

	// 2) Detect
	det, err := cli.Detect(ctx, url)
	if err != nil {
		log.Fatalf("Detect error: %v", err)
	}
	fmt.Println("Detection result:")
	fmt.Printf("  MIME:     %s\n", det.MIME)
	fmt.Printf("  Charset:  %s\n", det.Charset)
	fmt.Printf("  Encoding: %s\n", det.Encoding)
	fmt.Printf("  IsBinary: %v\n", det.IsBinary)
	fmt.Printf("  Title:    %s\n", det.Title)
	fmt.Println()

	// 3) ParseHTML
	parsed, err := cli.ParseHTML(res.Body)
	if err != nil {
		log.Fatalf("ParseHTML error: %v", err)
	}
	fmt.Println("Parsed HTML:")
	fmt.Printf("  Title: %s\n", parsed.Title)
	if len(parsed.Headings) > 0 {
		fmt.Printf("  First heading: [h%d] %s\n", parsed.Headings[0].Level, parsed.Headings[0].Text)
	}
	fmt.Printf("  Paragraphs: %d\n", len(parsed.Paragraphs))
	fmt.Printf("  Links:      %d\n", len(parsed.Links))
	fmt.Println()

	// 4) ExtractArticle
	article, err := cli.ExtractArticle(ctx, url)
	if err != nil {
		log.Fatalf("ExtractArticle error: %v", err)
	}
	fmt.Println("Extracted Article:")
	fmt.Printf("  Title:   %s\n", article.Title)
	fmt.Printf("  Byline:  %s\n", article.Byline)
	fmt.Printf("  Excerpt: %s\n", article.Excerpt)
	if len(article.Content) > 200 {
		fmt.Printf("  Content (first 200 chars): %s\n", article.Content[:200])
	} else {
		fmt.Printf("  Content: %s\n", article.Content)
	}
}
