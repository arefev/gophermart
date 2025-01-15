package trm

import (
	"context"
	"fmt"
)

type TrAction func(context.Context) error

type Transaction interface {
	Commit(context.Context) error
	Rollback(context.Context) error
	Begin(context.Context) (context.Context, error)
}

type trm struct {
	tr Transaction
}

func NewTrm(tr Transaction) *trm {
	return &trm{tr: tr}
}

func (trm *trm) Do(ctx context.Context, action TrAction) error {
	var err error
	ctx, err = trm.tr.Begin(ctx)
	if err != nil {
		return fmt.Errorf("trm begin fail: %w", err)
	}

	defer func() {
		if err := trm.tr.Rollback(ctx); err != nil {
			// TODO: handle error
		}
	}()

	if err := action(ctx); err != nil {
		return fmt.Errorf("trm action fail: %w", err)
	}

	err = trm.tr.Commit(ctx)
	if err != nil {
		return fmt.Errorf("trm commit fail: %w", err)
	}

	return nil
}