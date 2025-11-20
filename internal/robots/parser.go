// internal/robots/parser.go
package robots

import (
	"bufio"
	"bytes"
	"strings"
)

// Parse constructs a Robots structure from the given robots.txt bytes.
// The parser is intentionally simple but sufficient for Aether's goal
// of legal, respectful access.
func Parse(data []byte) *Robots {
	r := &Robots{}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var current *Group

	for scanner.Scan() {
		line := scanner.Text()
		// Strip comments.
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		field := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case "user-agent":
			if current == nil || len(current.Rules) > 0 {
				// start a new group
				current = &Group{}
				r.Groups = append(r.Groups, *current)
				current = &r.Groups[len(r.Groups)-1]
			}
			current.Agents = append(current.Agents, strings.ToLower(value))

		case "disallow":
			if current == nil {
				continue
			}
			current.Rules = append(current.Rules, Rule{
				Allow: false,
				Path:  value,
			})

		case "allow":
			if current == nil {
				continue
			}
			current.Rules = append(current.Rules, Rule{
				Allow: true,
				Path:  value,
			})
		default:
			// ignore other fields for now (crawl-delay, sitemap, etc.)
		}
	}

	return r
}
