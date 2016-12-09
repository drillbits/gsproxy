package main

import "context"

type private struct{}

var cfgKey private

// Config is a configuration
type Config struct {
	KeyFile string
}

// DefaultConfig returns default configuration.
func DefaultConfig() *Config {
	return &Config{}
}

// NewContext creates a new context.
func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, cfgKey, cfg)
}

// ConfigFromContext returns a configuration in context or default one.
func ConfigFromContext(ctx context.Context) *Config {
	if cfg, ok := ctx.Value(cfgKey).(*Config); ok && cfg != nil {
		return cfg
	}
	return DefaultConfig()
}
