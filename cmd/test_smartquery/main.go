// cmd/test_smartquery/main.go
//
// Test program for Aether SmartQuery classification and routing.

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

	tests := []string{
		"https://example.com",
		"Finland",
		"latest tech news",
		"python error: how to fix importerror",
		"rss hackernews",
		"github repo go-chi/chi",
		"documentation golang interfaces",
	}

	fmt.Println("=== SmartQuery Classification Tests ===")

	for _, q := range tests {
		p := cli.SmartQuery(q)

		fmt.Println("--------------------------------------------------")
		fmt.Printf("Query: %q\n", p.Query)
		fmt.Printf("Intent: %s\n", p.Intent)
		fmt.Printf("IsQuestion: %v\n", p.IsQuestion)
		fmt.Printf("HasURL: %v\n", p.HasURL)

		if len(p.PrimarySources) > 0 {
			fmt.Printf("Primary Sources: %v\n", p.PrimarySources)
		} else {
			fmt.Println("Primary Sources: <none>")
		}

		if len(p.FallbackSources) > 0 {
			fmt.Printf("Fallback Sources: %v\n", p.FallbackSources)
		} else {
			fmt.Println("Fallback Sources: <none>")
		}

		fmt.Printf("UseLookup: %v, UseSearchIndex: %v, UseOpenAPIs: %v, UseFeeds: %v, UsePlugins: %v\n",
			p.UseLookup, p.UseSearchIndex, p.UseOpenAPIs, p.UseFeeds, p.UsePlugins)
	}
}
