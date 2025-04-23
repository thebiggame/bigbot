// Package config provides the configuration state and utilities for working with the global bot configuration.
package config

import "log/slog"

// Type SecretString extends String with logging secrecy.
type SecretString string

// LogValue implements the slog.LogValuer interface.
func (SecretString) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}
