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

type Base struct {
	log *zap.Logger
}

func NewBase(log *zap.Logger) *Base {
	return &Base{log: log}
}

func (b *Base) findWithArgs(ctx context.Context, tx *sqlx.Tx, args map[string]any, query string, entity any) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("find with args: prepare named context fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			b.log.Warn("find with args: stmt close fail", zap.Error(err))
		}
	}()

	if err := stmt.GetContext(ctx, entity, args); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("find with args: get fail: %w", err)
		}
	}

	return nil
}

func (b *Base) createWithArgs(
	ctx context.Context,
	tx *sqlx.Tx,
	args map[string]any,
	query string,
) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("create with args: prepare named context fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			b.log.Warn("create with args: stmt close fail", zap.Error(err))
		}
	}()

	res, err := stmt.ExecContext(ctx, args)

	if err != nil {
		return nil, fmt.Errorf("create with args: exec query fail: %w", err)
	}

	return res, nil
}

func (b *Base) getWithArgs(
	ctx context.Context,
	tx *sqlx.Tx,
	args map[string]any,
	query string,
	list interface{},
) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create with args: prepare named context fail: %w", err)
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
