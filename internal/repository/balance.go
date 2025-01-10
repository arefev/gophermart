package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Balance struct {
	log *zap.Logger
	*Base
}

func NewBalance(log *zap.Logger) *Balance {
	return &Balance{
		log:  log,
		Base: NewBase(log),
	}
}

func (b *Balance) FindByUserID(tx *sqlx.Tx, userID int) (*model.Balance, bool) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	balance := model.Balance{}
	query := "SELECT id, user_id, current, withdrawn, created_at, updated_at FROM users_balance WHERE user_id = :user_id"
	arg := map[string]interface{}{"user_id": userID}

	ok, err := b.findWithArgs(ctx, tx, arg, query, &balance)
	if err != nil {
		b.log.Debug("find balance by user id: find with args fail: %w", zap.Error(err))
		return nil, false
	}

	return &balance, ok
}

func (b *Balance) UpdateByID(tx *sqlx.Tx, id int, current, withdrawn float64) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "UPDATE users_balance SET current = :current, withdrawn = :withdrawn WHERE id = :id"
	args := map[string]interface{}{
		"id":        id,
		"current":   current,
		"withdrawn": withdrawn,
	}

	if err := b.execWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("accrual by id fail: %w", err)
	}

	return nil
}
