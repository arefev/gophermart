package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const timeCancel = 15 * time.Second

type TxGetter interface {
	FromCtx(context.Context) (*sqlx.Tx, error)
}

type Base struct {
	log *zap.Logger
	tr  TxGetter
}

func NewBase(tr TxGetter, log *zap.Logger) *Base {
	return &Base{
		log: log,
		tr:  tr,
	}
}

func (b *Base) findWithArgs(ctx context.Context, args map[string]any, q string, entity any) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := b.prepare(ctx, q)
	if err != nil {
		return false, fmt.Errorf("exec with args: prepare fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			b.log.Warn("find with args: stmt close fail", zap.Error(err))
		}
	}()

	if err := stmt.GetContext(ctx, entity, args); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("find with args: get fail: %w", err)
	}

	return true, nil
}

func (b *Base) execWithArgs(ctx context.Context, args map[string]any, q string) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := b.prepare(ctx, q)
	if err != nil {
		return fmt.Errorf("exec with args: prepare fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			b.log.Warn("exec with args: stmt close fail", zap.Error(err))
		}
	}()

	_, err = stmt.ExecContext(ctx, args)
	if err != nil {
		return fmt.Errorf("exec with args: exec query fail: %w", err)
	}

	return nil
}

func (b *Base) getWithArgs(ctx context.Context, args map[string]any, q string, list any) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := b.prepare(ctx, q)
	if err != nil {
		return fmt.Errorf("exec with args: prepare fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			b.log.Warn("create with args: stmt close fail", zap.Error(err))
		}
	}()

	err = stmt.SelectContext(ctx, list, args)

	if err != nil {
		return fmt.Errorf("create with args: exec query fail: %w", err)
	}

	return nil
}

func (b *Base) prepare(ctx context.Context, q string) (*sqlx.NamedStmt, error) {
	tr, err := b.tr.FromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare from ctx fail: %w", err)
	}

	stmt, err := tr.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("prepare named context fail: %w", err)
	}

	return stmt, nil
}
