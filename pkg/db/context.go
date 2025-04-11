package db

import "context"

const ContextKeyDB = "db"

func FromContext(ctx context.Context) *DB {
	if db, ok := ctx.Value(ContextKeyDB).(*DB); ok {
		return db
	}
	return nil
}

func WithContext(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, ContextKeyDB, db)
}
