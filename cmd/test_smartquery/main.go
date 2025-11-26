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

		fmt.Printf("\nQuery: %q\nIntent: %s\nPrimarySources: %v\nFallback: %v\n",
			p.Query, p.Intent, p.PrimarySources, p.FallbackSources)
	}
}
