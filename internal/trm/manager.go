package trm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type TrAction func(context.Context) error

type Transaction interface {
	Commit(context.Context) error
	Rollback(context.Context) error
	Begin(context.Context) (context.Context, error)
}

type trm struct {
	tr  Transaction
	log *zap.Logger
}

func NewTrm(tr Transaction, log *zap.Logger) *trm {
	return &trm{
		tr:  tr,
		log: log,
	}
}

func (trm *trm) Do(ctx context.Context, action TrAction) error {
	var err error
	ctx, err = trm.tr.Begin(ctx)
	if err != nil {
		return fmt.Errorf("trm begin fail: %w", err)
	}

	defer func() {
		if err := trm.tr.Rollback(ctx); err != nil {
			trm.log.Error("trm rollback fail", zap.Error(err))
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
