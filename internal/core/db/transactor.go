package db

import (
	"context"
	"fmt"
)

type Transactor struct {
	pool *Pool
}

func NewTransactor(pool *Pool) *Transactor {
	return &Transactor{pool: pool}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := TxToContext(ctx, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
