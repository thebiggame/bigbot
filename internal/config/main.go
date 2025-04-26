// Package config provides the configuration state and utilities for working with the global bot configuration.
package config

import (
	"log/slog"
	"runtime/debug"
)

// Type SecretString extends String with logging secrecy.
type SecretString string

// LogValue implements the slog.LogValuer interface.
func (SecretString) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}

var AppVersion = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				// Return the first 7 characters of the commit hash, like a short hash.
				return setting.Value[0:7]
			}
		}
	}
	return "UNKNOWN"
}()
