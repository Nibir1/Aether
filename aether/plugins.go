// aether/plugins.go
//
// Public plugin-registration API for Aether.
// This file bridges plugin authors (github.com/Nibir1/Aether/plugins)
// with the internal Client mechanics.
//
// Plugin categories (Stage 13):
//   • Source plugins      — legal/public data sources
//   • Transform plugins   — post-normalization enrichers
//   • Display plugins     — alternative renderers
//
// The Client owns one registry instance, created during NewClient().

package aether

import (
	"fmt"

	"github.com/Nibir1/Aether/plugins"
)

// initPlugins is called internally during NewClient() to bind
// a new plugin registry owned by the Client.
func (c *Client) initPlugins() {
	c.plugins = plugins.NewRegistry()
}

// RegisterSourcePlugin registers a SourcePlugin.
//
// Strict naming:
// If another plugin with the same name is already registered,
// an error is returned by the registry.
func (c *Client) RegisterSourcePlugin(p plugins.SourcePlugin) error {
	if c == nil {
		return fmt.Errorf("aether: nil client")
	}
	if c.plugins == nil {
		return fmt.Errorf("aether: plugin registry not initialized")
	}
	return c.plugins.RegisterSource(p)
}

// RegisterTransformPlugin registers a TransformPlugin.
// These run after normalization in the Search pipeline.
func (c *Client) RegisterTransformPlugin(p plugins.TransformPlugin) error {
	if c == nil {
		return fmt.Errorf("aether: nil client")
	}
	if c.plugins == nil {
		return fmt.Errorf("aether: plugin registry not initialized")
	}
	return c.plugins.RegisterTransform(p)
}

// RegisterDisplayPlugin registers a DisplayPlugin.
// These produce non-Markdown formats (HTML, ANSI, PDF, …).
func (c *Client) RegisterDisplayPlugin(p plugins.DisplayPlugin) error {
	if c == nil {
		return fmt.Errorf("aether: nil client")
	}
	if c.plugins == nil {
		return fmt.Errorf("aether: plugin registry not initialized")
	}
	return c.plugins.RegisterDisplay(p)
}
