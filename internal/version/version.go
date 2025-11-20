// internal/version/version.go
//
// Package version contains the Aether library version string.
// This is kept in an internal package so that the public API can
// expose it in a controlled manner via aether.Version().
package version

// AetherVersion is the current version of the Aether library.
//
// During early development this may be a "-dev" version. For tagged
// releases it should follow semantic versioning, e.g. "v1.0.0".
const AetherVersion = "v0.1.0-dev"
