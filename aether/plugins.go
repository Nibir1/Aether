// aether/plugins.go
//
// This file exposes the public plugin-registration API for Aether.
// It acts as the bridge between the public plugin interfaces in
// github.com/Nibir1/Aether/plugins and the internal Client mechanics.
//
// Stage 13 introduces three plugin categories:
//
//   • Source plugins      — new legal/public data sources
//   • Transform plugins   — post-normalization enrichers
//   • Display plugins     — alternative output/format renderers
//
// The registry itself lives in the public plugins package so that
// plugin authors can reference types and error semantics without
// importing internal Aether packages.
//
// The Client owns a single plugin registry instance, created during
// NewClient(), and all registration functions below delegate to it.

package aether

import (
	"fmt"

	"github.com/Nibir1/Aether/plugins"
)

// initPlugins is called internally during NewClient() to bind a new registry.
// This keeps plugin ownership private to the Client.
func (c *Client) initPlugins() {
	c.plugins = plugins.NewRegistry()
}

// RegisterSourcePlugin registers a public SourcePlugin with the Client.
//
// Usage:
//
//	cli := aether.NewClient(...)
//	err := cli.RegisterSourcePlugin(myHNPlugin)
//
// Strict naming rule:
//
//	If another plugin with the same name already exists, an error is returned.
func (c *Client) RegisterSourcePlugin(p plugins.SourcePlugin) error {
	if c == nil {
		return fmt.Errorf("aether.Client is nil")
	}
	if c.plugins == nil {
		return fmt.Errorf("plugin registry not initialized")
	}
	return c.plugins.RegisterSource(p)
}

// RegisterTransformPlugin registers a transform plugin.
//
// Transform plugins take a normalized Document and return a modified/enriched
// version. They run after built-in normalization in the Search pipeline.
func (c *Client) RegisterTransformPlugin(p plugins.TransformPlugin) error {
	if c == nil {
		return fmt.Errorf("aether.Client is nil")
	}
	if c.plugins == nil {
		return fmt.Errorf("plugin registry not initialized")
	}
	return c.plugins.RegisterTransform(p)
}

// RegisterDisplayPlugin registers a display plugin.
//
// Display plugins output formats other than Markdown (e.g., HTML, ANSI, PDF).
// Later stages will integrate these into Aether’s rendering subsystem.
func (c *Client) RegisterDisplayPlugin(p plugins.DisplayPlugin) error {
	if c == nil {
		return fmt.Errorf("aether.Client is nil")
	}
	if c.plugins == nil {
		return fmt.Errorf("plugin registry not initialized")
	}
	return c.plugins.RegisterDisplay(p)
}

// listSourcePlugins returns the registered source plugin names.
// This is used internally by SmartQuery routing.
func (c *Client) listSourcePlugins() []string {
	if c == nil || c.plugins == nil {
		return nil
	}
	return c.plugins.ListSources()
}

// listTransformPlugins returns registered Transform plugins in stable order.
// Used internally after normalization.
func (c *Client) listTransformPlugins() []string {
	if c == nil || c.plugins == nil {
		return nil
	}
	return c.plugins.ListTransforms()
}

// listDisplayPlugins returns all display-plugin format identifiers.
// Used by future multi-format rendering APIs.
func (c *Client) listDisplayPlugins() []string {
	if c == nil || c.plugins == nil {
		return nil
	}
	return c.plugins.ListDisplays()
}
