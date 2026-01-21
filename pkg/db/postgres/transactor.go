package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBeginTransactionFailed    = errors.New("failed to begin transaction")
	ErrCouldNotCommitTransaction = errors.New("could not commit transaction")
)

type Transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) (*Transactor, QuerierProvider) {
	t := &Transactor{pool: pool}

	qp := func(ctx context.Context) Querier {
		if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok && tx != nil {
			return tx
		}
		return pool
	}

	return t, qp
}

func (t *Transactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return ErrBeginTransactionFailed
	}

	txCtx := txToContext(ctx, tx)

	// If callback fails: rollback and return the callback error.
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	// If commit fails: attempt rollback and return commit error.
	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return ErrCouldNotCommitTransaction
	}

	return nil
}

type txCtxKey struct{}

// txToContext stores the transaction in context (your existing approach).
func txToContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// querierFromContext returns tx if present, otherwise falls back to the pool.
func querierFromContext(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return pool
}
