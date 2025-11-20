// plugins/registry.go
//
// This file implements the public-facing plugin registry for Aether.
// It defines a thread-safe structure used by the aether.Client to
// organize and invoke user-registered plugins.
//
// NOTE:
// The registry is intentionally defined in the *public* plugins package
// so plugin authors can reference registry behavior and errors, but
// the registry is instantiated and owned internally by the Aether client.
// External users do NOT create registries manually.
//
// Registry rules:
//   • Plugin names must be unique (strict mode).
//   • Registration is thread-safe.
//   • Lookups are stable and deterministic.
//   • No plugin may override another unless Aether explicitly
//     enables override mode in a future extension.
//
// This file contains no imports from the aether package to avoid
// circular dependencies.

package plugins

import (
	"errors"
	"fmt"
	"sync"
)

// Registry holds all registered plugins.
// It is safe for concurrent use by the Aether client and plugins.
//
// Aether creates one Registry per client instance.
type Registry struct {
	mu sync.RWMutex

	sources    map[string]SourcePlugin
	transforms map[string]TransformPlugin
	displays   map[string]DisplayPlugin
}

// NewRegistry constructs an empty, thread-safe plugin registry.
// This function is used internally by the Aether client.
func NewRegistry() *Registry {
	return &Registry{
		sources:    make(map[string]SourcePlugin),
		transforms: make(map[string]TransformPlugin),
		displays:   make(map[string]DisplayPlugin),
	}
}

// errAlreadyRegistered is returned when a plugin name is already registered.
var errAlreadyRegistered = errors.New("plugin with this name already registered")

//
// ─────────────────────────────────────────────
//          SOURCE PLUGIN REGISTRATION
// ─────────────────────────────────────────────
//

// RegisterSource registers a SourcePlugin. Names must be unique.
func (r *Registry) RegisterSource(p SourcePlugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil SourcePlugin")
	}
	name := p.Name()
	if name == "" {
		return fmt.Errorf("SourcePlugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sources[name]; exists {
		return fmt.Errorf("%w: %s", errAlreadyRegistered, name)
	}

	r.sources[name] = p
	return nil
}

// GetSource returns a SourcePlugin by name, or nil if not found.
func (r *Registry) GetSource(name string) SourcePlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sources[name]
}

// ListSources returns a sorted list of SourcePlugin names.
func (r *Registry) ListSources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.sources))
	for n := range r.sources {
		out = append(out, n)
	}
	sortStrings(out)
	return out
}

//
// ─────────────────────────────────────────────
//        TRANSFORM PLUGIN REGISTRATION
// ─────────────────────────────────────────────
//

// RegisterTransform registers a TransformPlugin. Names must be unique.
func (r *Registry) RegisterTransform(p TransformPlugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil TransformPlugin")
	}
	name := p.Name()
	if name == "" {
		return fmt.Errorf("TransformPlugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transforms[name]; exists {
		return fmt.Errorf("%w: %s", errAlreadyRegistered, name)
	}

	r.transforms[name] = p
	return nil
}

// GetTransform returns a TransformPlugin by name.
func (r *Registry) GetTransform(name string) TransformPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transforms[name]
}

// ListTransforms returns a sorted list of transform plugin names.
func (r *Registry) ListTransforms() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.transforms))
	for n := range r.transforms {
		out = append(out, n)
	}
	sortStrings(out)
	return out
}

//
// ─────────────────────────────────────────────
//         DISPLAY PLUGIN REGISTRATION
// ─────────────────────────────────────────────
//

// RegisterDisplay registers a DisplayPlugin. Names must be unique.
func (r *Registry) RegisterDisplay(p DisplayPlugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil DisplayPlugin")
	}
	name := p.Name()
	if name == "" {
		return fmt.Errorf("DisplayPlugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.displays[name]; exists {
		return fmt.Errorf("%w: %s", errAlreadyRegistered, name)
	}

	r.displays[name] = p
	return nil
}

// GetDisplay returns a DisplayPlugin by name.
func (r *Registry) GetDisplay(name string) DisplayPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.displays[name]
}

// ListDisplays returns a sorted list of display plugin names.
func (r *Registry) ListDisplays() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.displays))
	for n := range r.displays {
		out = append(out, n)
	}
	sortStrings(out)
	return out
}

//
// ─────────────────────────────────────────────
//                 UTILITIES
// ─────────────────────────────────────────────
//

// sortStrings performs an in-place lexicographic sort.
// We cannot import "sort" here because it would be unnecessary overhead,
// and implementing a tiny insertion sort keeps this minimal and predictable.
func sortStrings(s []string) {
	n := len(s)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 && s[j] < s[j-1] {
			s[j], s[j-1] = s[j-1], s[j]
			j--
		}
	}
}
