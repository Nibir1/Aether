// internal/toon/bton.go
//
// BTON — Binary TOON v1
//
// A compact binary encoding for TOON tokens, used for:
//   • High-speed caching
//   • Inter-agent RPC
//   • Space-constrained storage
//
// Encoding:
//
//   MAGIC: "BTON\x00"
//   Document:
//       [sourceURLLen][sourceURL]
//       [kindLen][kind]
//       [titleLen][title]
//       [excerptLen][excerpt]
//       [attrCount][keyLen][key][valLen][val]...
//       [tokenCount]
//           token{
//             typeByte
//             roleLen, role
//             textLen, text
//             attrCount, [keyLen,key,valLen,val]...
//           }
//
// Token types are mapped to small bytes via encodeTokenType/decodeTokenType.

package toon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/Nibir1/Aether/internal/model"
)

const btonMagic = "BTON\x00"

var (
	errInvalidBTON = errors.New("aether/toon: invalid BTON stream")
)

const (
	btonTypeUnknown  = 0
	btonTypeText     = 1
	btonTypeHeading  = 2
	btonTypeSectionS = 3
	btonTypeSectionE = 4
	btonTypeMeta     = 5
	btonTypeDocInfo  = 6
	btonTypeTitle    = 7
	btonTypeExcerpt  = 8
)

func encodeTokenType(t TokenType) byte {
	switch t {
	case TokenText:
		return btonTypeText
	case TokenHeading:
		return btonTypeHeading
	case TokenSectionStart:
		return btonTypeSectionS
	case TokenSectionEnd:
		return btonTypeSectionE
	case TokenMeta:
		return btonTypeMeta
	case TokenDocumentInfo:
		return btonTypeDocInfo
	case TokenTitle:
		return btonTypeTitle
	case TokenExcerpt:
		return btonTypeExcerpt
	default:
		return btonTypeUnknown
	}
}

func decodeTokenType(b byte) TokenType {
	switch b {
	case btonTypeText:
		return TokenText
	case btonTypeHeading:
		return TokenHeading
	case btonTypeSectionS:
		return TokenSectionStart
	case btonTypeSectionE:
		return TokenSectionEnd
	case btonTypeMeta:
		return TokenMeta
	case btonTypeDocInfo:
		return TokenDocumentInfo
	case btonTypeTitle:
		return TokenTitle
	case btonTypeExcerpt:
		return TokenExcerpt
	default:
		return TokenText
	}
}

//
// ─────────────────────────────────────────────
//                   HELPERS
// ─────────────────────────────────────────────
//

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	if len(s) == 0 {
		return nil
	}
	_, err := w.Write([]byte(s))
	return err
}

func readString(r io.Reader) (string, error) {
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

//
// ─────────────────────────────────────────────
//                 ENCODE BTON
// ─────────────────────────────────────────────
//

// EncodeBTON encodes a TOON Document into binary BTON format.
func EncodeBTON(doc *Document) ([]byte, error) {
	if doc == nil {
		doc = &Document{}
	}

	buf := &bytes.Buffer{}

	// Magic header
	if _, err := buf.WriteString(btonMagic); err != nil {
		return nil, err
	}

	// Basic fields
	if err := writeString(buf, doc.SourceURL); err != nil {
		return nil, err
	}
	if err := writeString(buf, string(doc.Kind)); err != nil {
		return nil, err
	}
	if err := writeString(buf, doc.Title); err != nil {
		return nil, err
	}
	if err := writeString(buf, doc.Excerpt); err != nil {
		return nil, err
	}

	// Attributes
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(doc.Attributes))); err != nil {
		return nil, err
	}
	for k, v := range doc.Attributes {
		if err := writeString(buf, k); err != nil {
			return nil, err
		}
		if err := writeString(buf, v); err != nil {
			return nil, err
		}
	}

	// Token stream
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(doc.Tokens))); err != nil {
		return nil, err
	}

	for _, tok := range doc.Tokens {
		// Token type as a single byte
		if err := buf.WriteByte(encodeTokenType(tok.Type)); err != nil {
			return nil, err
		}

		// Role + Text
		if err := writeString(buf, tok.Role); err != nil {
			return nil, err
		}
		if err := writeString(buf, tok.Text); err != nil {
			return nil, err
		}

		// Attributes
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(tok.Attrs))); err != nil {
			return nil, err
		}
		for k, v := range tok.Attrs {
			if err := writeString(buf, k); err != nil {
				return nil, err
			}
			if err := writeString(buf, v); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

//
// ─────────────────────────────────────────────
//                 DECODE BTON
// ─────────────────────────────────────────────
//

// DecodeBTON parses a BTON binary stream back into a TOON Document.
func DecodeBTON(b []byte) (*Document, error) {
	r := bytes.NewReader(b)

	// Validate magic header
	magic := make([]byte, len(btonMagic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if string(magic) != btonMagic {
		return nil, errInvalidBTON
	}

	doc := &Document{}

	var err error
	if doc.SourceURL, err = readString(r); err != nil {
		return nil, err
	}
	kindStr, err := readString(r)
	if err != nil {
		return nil, err
	}
	doc.Kind = model.DocumentKind(kindStr)

	if doc.Title, err = readString(r); err != nil {
		return nil, err
	}
	if doc.Excerpt, err = readString(r); err != nil {
		return nil, err
	}

	// Attributes
	var attrCount uint32
	if err := binary.Read(r, binary.LittleEndian, &attrCount); err != nil {
		return nil, err
	}
	if attrCount > 0 {
		doc.Attributes = make(map[string]string, attrCount)
		for i := 0; i < int(attrCount); i++ {
			k, err := readString(r)
			if err != nil {
				return nil, err
			}
			v, err := readString(r)
			if err != nil {
				return nil, err
			}
			doc.Attributes[k] = v
		}
	}

	// Tokens
	var tokenCount uint32
	if err := binary.Read(r, binary.LittleEndian, &tokenCount); err != nil {
		return nil, err
	}
	doc.Tokens = make([]Token, tokenCount)

	for i := 0; i < int(tokenCount); i++ {
		var t Token

		typeByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		t.Type = decodeTokenType(typeByte)

		if t.Role, err = readString(r); err != nil {
			return nil, err
		}
		if t.Text, err = readString(r); err != nil {
			return nil, err
		}

		var n uint32
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return nil, err
		}
		if n > 0 {
			t.Attrs = make(map[string]string, n)
			for j := 0; j < int(n); j++ {
				k, err := readString(r)
				if err != nil {
					return nil, err
				}
				v, err := readString(r)
				if err != nil {
					return nil, err
				}
				t.Attrs[k] = v
			}
		}

		doc.Tokens[i] = t
	}

	return doc, nil
}
