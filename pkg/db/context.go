package db

import "context"

var ContextKey = struct{ string }{"db"}

func FromContext(ctx context.Context) *DB {
	if db, ok := ctx.Value(ContextKey).(*DB); ok {
		return db
	}
	return nil
}

func WithContext(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, ContextKey, db)
}
