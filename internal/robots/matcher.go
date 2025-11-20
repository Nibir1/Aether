// internal/robots/matcher.go
package robots

import "strings"

// Allowed reports whether the given userAgent is allowed to access
// the specified path according to the robots.txt rules.
//
// It implements the "longest match" rule: the rule with the longest
// matching path prefix wins. If no rules match, access is allowed.
func (r *Robots) Allowed(userAgent, path string) bool {
	if r == nil {
		return true
	}
	ua := strings.ToLower(userAgent)

	// Collect groups matching this user agent.
	matchingGroups := make([]Group, 0, len(r.Groups))
	for _, g := range r.Groups {
		if groupMatches(g, ua) {
			matchingGroups = append(matchingGroups, g)
		}
	}

	// If no explicit group matches, consider "*" groups.
	if len(matchingGroups) == 0 {
		for _, g := range r.Groups {
			for _, a := range g.Agents {
				if a == "*" {
					matchingGroups = append(matchingGroups, g)
					break
				}
			}
		}
	}

	if len(matchingGroups) == 0 {
		return true
	}

	// Longest-path rule resolution.
	var (
		bestRule   *Rule
		bestLength int
	)
	for _, g := range matchingGroups {
		for i := range g.Rules {
			rule := &g.Rules[i]
			if rule.Path == "" {
				// Empty disallow path means "allow all".
				continue
			}
			if strings.HasPrefix(path, rule.Path) {
				if len(rule.Path) > bestLength {
					bestRule = rule
					bestLength = len(rule.Path)
				}
			}
		}
	}

	if bestRule == nil {
		return true
	}
	return bestRule.Allow
}

func groupMatches(g Group, ua string) bool {
	for _, a := range g.Agents {
		if a == "*" || a == ua {
			return true
		}
		// Some robots files use partial matches; we support prefix matching.
		if strings.Contains(ua, a) && a != "" {
			return true
		}
	}
	return false
}
