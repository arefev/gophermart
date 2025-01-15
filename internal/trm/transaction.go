package trm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type tr struct {
	db *sqlx.DB
}

func NewTr(db *sqlx.DB) *tr {
    return &tr{db: db}
}

func (tr *tr) Begin(ctx context.Context) (context.Context, error) {
	tx, err := tr.db.Beginx()
	if err != nil {
		return ctx, fmt.Errorf("transaction begin fail: %w", err)
	}

	return context.WithValue(ctx, sqlx.Tx{}, tx), nil
}

func (tr *tr) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(sqlx.Tx{}).(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction commit fail: %w", err)
	}

	return nil
}

func (tr *tr) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(sqlx.Tx{}).(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	
	if err := tx.Rollback(); err != nil {
		if !errors.Is(err, sql.ErrTxDone) {
			return fmt.Errorf("transaction rollback fail: %w", err)
		}
	}

	return nil
}

func (tr *tr) FromCtx(ctx context.Context) (*sqlx.Tx, error) {
	tx, ok := ctx.Value(sqlx.Tx{}).(*sqlx.Tx)
    if !ok {
        return nil, fmt.Errorf("tx not found in context")
    }

    return tx, nil
}

