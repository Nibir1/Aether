// aether/detect.go
//
// Public content-detection API for Aether. This is used internally by
// Search and SmartQuery routing but is also available for direct use.

package aether

import (
	"context"

	idetect "github.com/Nibir1/Aether/internal/detect"
	ihtml "github.com/Nibir1/Aether/internal/html"
)

// DetectionResult is the public type returned to callers.
type DetectionResult struct {
	URL       string
	RawType   string
	SubType   string
	MIME      string
	Charset   string
	Encoding  string
	IsBinary  bool
	Metadata  map[string]string
	Title     string
	Canonical string
}

// Detect runs a full fetch (robots.txt-compliant), sniffs content type,
// parses HTML only when needed, and returns a structured detection result.
func (c *Client) Detect(ctx context.Context, url string) (*DetectionResult, error) {
	if c == nil {
		return nil, ErrNilClient
	}

	// Step 0: full legal fetch
	res, err := c.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	// Step 1: MIME + heuristic detection
	dr := idetect.Detect(res.Body, res.Header.Clone())

	out := &DetectionResult{
		URL:      url,
		RawType:  string(dr.RawType),
		SubType:  string(dr.SubType),
		MIME:     dr.MIME,
		Charset:  dr.Charset,
		Encoding: dr.Encoding,
		IsBinary: dr.IsBinary,
		Metadata: map[string]string{},
	}

	// Step 2: For HTML, extract title, description, canonical URL, etc.
	if dr.RawType == idetect.TypeHTML {
		doc, err := ihtml.ParseDocument(res.Body)
		if err == nil {
			meta := idetect.ExtractBasicMeta(doc)
			out.Metadata = meta
			out.Title = meta["title"]
			out.Canonical = meta["canonical_url"]
		}
	}

	return out, nil
}
