package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

type storeContextKey int

// ContextKeys for a store context
var (
	connectionContextKey = storeContextKey(1)
)

// Errors
var (
	// ErrNoConnectionInConfig will occure when the connection
	// is not available in the context.
	ErrNoConnectionInConfig = errors.New("connection missing in context")
)

// ContextWithConnection will return a child context
// with a value for connection.
func ContextWithConnection(
	ctx context.Context,
	conn *pgxpool.Conn,
) context.Context {
	return context.WithValue(ctx, connectionContextKey, conn)
}

// ConnectionFromContext will retrieve the connection.
func ConnectionFromContext(ctx context.Context) *pgxpool.Conn {
	conn, ok := ctx.Value(connectionContextKey).(*pgxpool.Conn)
	if !ok {
		panic(ErrNoConnectionInConfig)
	}
	if conn == nil {
		panic(ErrNoConnectionInConfig)
	}
	return conn
}
