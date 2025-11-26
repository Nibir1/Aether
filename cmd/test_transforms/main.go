package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Nibir1/Aether/aether"
	"github.com/Nibir1/Aether/plugins"
)

// -----------------------------
// Dummy Transform Plugin
// -----------------------------
type UppercaseTitleTransform struct{}

func (UppercaseTitleTransform) Name() string        { return "uppercase_title" }
func (UppercaseTitleTransform) Description() string { return "uppercase title for testing" }

func (UppercaseTitleTransform) Apply(ctx context.Context, doc *plugins.Document) (*plugins.Document, error) {
	// Modify title and return
	doc.Title = strings.ToUpper(doc.Title)
	return doc, nil
}

func main() {
	cli, err := aether.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Register transform plugin
	err = cli.RegisterTransformPlugin(UppercaseTitleTransform{})
	if err != nil {
		log.Fatal(err)
	}

	sr, err := cli.Search(context.Background(), "Finland")
	if err != nil {
		log.Fatal(err)
	}

	normalized := cli.NormalizeSearchResult(sr)

	fmt.Println("Title after TransformPlugins:")
	fmt.Println("->", normalized.Title)
}
