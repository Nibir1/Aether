// internal/detect/detect.go
//
// Package detect provides Aether’s internal content-type detection.
// It determines whether the content is HTML, JSON, XML, RSS/Atom,
// plaintext, PDF, or binary. It also attempts to classify HTML pages
// such as article vs homepage vs documentation.
//
// Detection is lightweight and runs before heavy operations like
// parsing or extraction.

package detect

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// Type represents Aether’s internal content classification.
type Type string

const (
	TypeUnknown  Type = "unknown"
	TypeHTML     Type = "html"
	TypeJSON     Type = "json"
	TypeXML      Type = "xml"
	TypeRSS      Type = "rss"
	TypePDF      Type = "pdf"
	TypeText     Type = "text"
	TypeImage    Type = "image"
	TypeBinary   Type = "binary"
	TypeArticle  Type = "article"
	TypeHomepage Type = "homepage"
	TypeDocs     Type = "docs"
)

// Result is the internal detection output.
type Result struct {
	RawType  Type   // basic type from content sniffing
	SubType  Type   // deeper classification (article/docs)
	MIME     string // HTTP Content-Type header
	Metadata map[string]string
	Charset  string
	Encoding string
	IsBinary bool
}

// Detect performs content detection on HTTP response body and headers.
func Detect(body []byte, headers http.Header) *Result {
	mime := strings.ToLower(headers.Get("Content-Type"))
	r := &Result{
		MIME:     mime,
		Metadata: map[string]string{},
	}

	// --- MIME-based detection ---
	switch {
	case strings.Contains(mime, "html"):
		r.RawType = TypeHTML
	case strings.Contains(mime, "json"):
		r.RawType = TypeJSON
	case strings.Contains(mime, "xml"):
		r.RawType = TypeXML
	case strings.Contains(mime, "rss"), strings.Contains(mime, "atom+xml"):
		r.RawType = TypeRSS
	case strings.Contains(mime, "pdf"):
		r.RawType = TypePDF
	case strings.Contains(mime, "text/plain"):
		r.RawType = TypeText
	case strings.Contains(mime, "image/"):
		r.RawType = TypeImage
		r.IsBinary = true
	default:
		// Try fallback content sniffing
		r.RawType = sniff(body)
	}

	// Subtype detection only for HTML
	if r.RawType == TypeHTML {
		r.SubType = classifyHTML(body)
	}

	return r
}

// sniff guesses content type from body content.
func sniff(body []byte) Type {
	b := bytes.TrimSpace(body)
	if len(b) == 0 {
		return TypeUnknown
	}

	// JSON object/array?
	if bytes.HasPrefix(b, []byte("{")) || bytes.HasPrefix(b, []byte("[")) {
		var js json.RawMessage
		if json.Unmarshal(b, &js) == nil {
			return TypeJSON
		}
	}

	// HTML?
	s := strings.ToLower(string(b[:64]))
	if strings.Contains(s, "<!doctype html") || strings.Contains(s, "<html") {
		return TypeHTML
	}

	// XML?
	if bytes.HasPrefix(b, []byte("<?xml")) {
		return TypeXML
	}

	return TypeBinary
}

// classifyHTML runs very light heuristics for article/home/docs.
func classifyHTML(body []byte) Type {
	l := strings.ToLower(string(body))

	switch {
	case strings.Contains(l, "<article"):
		return TypeArticle
	case strings.Contains(l, "documentation") ||
		strings.Contains(l, "docs") ||
		strings.Contains(l, "api reference"):
		return TypeDocs
	case strings.Contains(l, "<nav") &&
		strings.Contains(l, "<main") &&
		strings.Contains(l, "<footer"):
		return TypeHomepage
	default:
		return TypeUnknown
	}
}
