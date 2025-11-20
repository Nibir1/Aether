// internal/rss/parser.go
//
// Robust XML-based parsing for RSS 2.0, RSS 1.0, and Atom feeds.
// Aether sniff-detects the feed type and normalizes them into a
// unified Feed struct.

package rss

import (
	"bytes"
	"encoding/xml"
	"strings"
	"time"
)

// --- Atom Structures ---

type atomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Link    []struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Entries []struct {
		Title     string `xml:"title"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
		ID        string `xml:"id"`
		Updated   string `xml:"updated"`
		Published string `xml:"published"`
		Author    struct {
			Name string `xml:"name"`
		} `xml:"author"`
		Links []struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

// --- RSS 2.0 Structures ---

type rss2Feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		LastBuild   string `xml:"lastBuildDate"`
		Items       []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			Content     string `xml:"encoded"`
			Author      string `xml:"author"`
			PubDate     string `xml:"pubDate"`
			GUID        string `xml:"guid"`
		} `xml:"item"`
	} `xml:"channel"`
}

// --- RSS 1.0 / RDF Structures ---

type rss1Feed struct {
	XMLName xml.Name `xml:"RDF"`
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
	} `xml:"channel"`
	Items []struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
	} `xml:"item"`
}

// Parse parses raw XML into a unified Feed structure.
func Parse(data []byte) (*Feed, error) {
	trim := bytes.TrimSpace(data)
	lower := strings.ToLower(string(trim[:64]))

	switch {
	case strings.Contains(lower, "<feed"):
		return parseAtom(data)
	case strings.Contains(lower, "<rss"):
		return parseRSS2(data)
	case strings.Contains(lower, "<rdf"):
		return parseRSS1(data)
	default:
		// Try RSS2 fallback
		if f, err := parseRSS2(data); err == nil {
			return f, nil
		}
		// Try Atom fallback
		if f, err := parseAtom(data); err == nil {
			return f, nil
		}
		// Try RSS1 fallback
		if f, err := parseRSS1(data); err == nil {
			return f, nil
		}
	}

	return nil, ErrUnknownFeed
}

var ErrUnknownFeed = &FeedError{"unknown or unsupported RSS/Atom format"}

// FeedError is an implementation of error for feed parsing.
type FeedError struct {
	Msg string
}

func (e *FeedError) Error() string { return e.Msg }

func parseAtom(data []byte) (*Feed, error) {
	var a atomFeed
	if err := xml.Unmarshal(data, &a); err != nil {
		return nil, err
	}

	f := &Feed{
		Title:   a.Title,
		Updated: parseTime(a.Updated),
	}

	if len(a.Link) > 0 {
		f.Link = a.Link[0].Href
	}

	for _, e := range a.Entries {
		link := ""
		if len(e.Links) > 0 {
			link = e.Links[0].Href
		}
		f.Items = append(f.Items, Item{
			Title:       e.Title,
			Link:        link,
			Description: e.Summary,
			Content:     e.Content,
			Author:      e.Author.Name,
			Published:   parseTime(e.Published),
			Updated:     parseTime(e.Updated),
			GUID:        e.ID,
		})
	}

	return f, nil
}

func parseRSS2(data []byte) (*Feed, error) {
	var r rss2Feed
	if err := xml.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	c := r.Channel
	f := &Feed{
		Title:       c.Title,
		Description: c.Description,
		Link:        c.Link,
		Updated:     parseTime(c.LastBuild),
	}

	for _, it := range c.Items {
		content := it.Content
		if content == "" {
			content = it.Description
		}
		f.Items = append(f.Items, Item{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
			Content:     content,
			Author:      it.Author,
			Published:   parseTime(it.PubDate),
			GUID:        it.GUID,
		})
	}

	return f, nil
}

func parseRSS1(data []byte) (*Feed, error) {
	var r rss1Feed
	if err := xml.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	f := &Feed{
		Title:       r.Channel.Title,
		Description: r.Channel.Description,
		Link:        r.Channel.Link,
	}

	for _, it := range r.Items {
		f.Items = append(f.Items, Item{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
		})
	}

	return f, nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC1123Z, s)
	if !t.IsZero() {
		return t
	}
	t, _ = time.Parse(time.RFC1123, s)
	if !t.IsZero() {
		return t
	}
	t, _ = time.Parse(time.RFC3339, s)
	return t
}
