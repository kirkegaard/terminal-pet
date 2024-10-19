package config

import "context"

var ContextKey = struct{ string }{"config"}

func WithContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ContextKey, cfg)
}

func FromContext(ctx context.Context) *Config {
	if c, ok := ctx.Value(ContextKey).(*Config); ok {
		return c
	}

	return nil
}
