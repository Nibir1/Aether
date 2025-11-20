// internal/openapi/github.go
//
// GitHub README integration using the raw content endpoint:
//   https://raw.githubusercontent.com/{owner}/{repo}/{ref}/README.md
//
// This avoids any authentication requirement and is sufficient for
// retrieving human-readable documentation for many repositories.

package openapi

import (
	"context"
	"fmt"
	"strings"
)

// GitHubReadme represents a normalized GitHub README document.
type GitHubReadme struct {
	Owner   string
	Repo    string
	Ref     string
	URL     string
	Content string
}

// GitHubReadme fetches the README.md file for the given repository.
// ref may be empty, in which case \"main\" is assumed.
func (c *Client) GitHubReadme(ctx context.Context, owner, repo, ref string) (*GitHubReadme, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	ref = strings.TrimSpace(ref)

	if owner == "" || repo == "" {
		return nil, nil
	}
	if ref == "" {
		ref = "main"
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md",
		owner, repo, ref)

	body, _, err := c.getText(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	return &GitHubReadme{
		Owner:   owner,
		Repo:    repo,
		Ref:     ref,
		URL:     rawURL,
		Content: string(body),
	}, nil
}
