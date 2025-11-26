package main

import (
	"fmt"
	"log"

	"github.com/Nibir1/Aether/aether"
)

func main() {
	cli, err := aether.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Construct a fake SearchResult with all three layers
	sr := &aether.SearchResult{
		Plan: aether.SearchPlan{Intent: "test"},
		PrimaryDocument: &aether.SearchDocument{
			URL:     "https://example.com",
			Title:   "Base Title",
			Excerpt: "Primary Excerpt",
			Content: "Primary Content",
			Metadata: map[string]string{
				"source": "primary",
			},
			Kind: aether.SearchDocumentKindHTML,
		},
		Article: &aether.Article{
			Title:   "Article Title",
			Content: "Article Body Content",
			Meta: map[string]string{
				"article_meta": "yes",
			},
		},
		Feed: &aether.Feed{
			Items: []aether.FeedItem{
				{
					Title: "Feed Item 1",
					Link:  "https://example.com/f1",
				},
			},
		},
	}

	norm := cli.NormalizeSearchResult(sr)

	fmt.Println("\n=== Normalized Merge Test ===")
	fmt.Println("Kind:", norm.Kind)
	fmt.Println("Title:", norm.Title)
	fmt.Println("Excerpt:", norm.Excerpt)
	fmt.Println("Content:", norm.Content)
	fmt.Println("Sections:", len(norm.Sections))
	fmt.Println("Metadata:", norm.Metadata)
}
