package store

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type storeContextKey int

// Context keys
var (
	txContextKey = storeContextKey(1)
)

// ContextWithTransaction adds a transaction to a context
func ContextWithTransaction(ctx context.Context, tx *pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}

// MustTransactionFromContext retrievs a transaction.
// If the transaction is missing we panic.
func MustTransactionFromContext(ctx context.Context) *pgx.Tx {
	tx, ok := ctx.Value(txContextKey).(*pgx.Tx)
	if !ok {
		panic("context requires a transaction")
	}
	return tx
}
