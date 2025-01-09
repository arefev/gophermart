package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Order struct {
	log *zap.Logger
	*Base
}

func NewOrder(log *zap.Logger) *Order {
	return &Order{
		log:  log,
		Base: NewBase(log),
	}
}

func (o *Order) FindByNumber(tx *sqlx.Tx, number string) *model.Order {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	order := model.Order{}
	args := map[string]any{"number": number}
	query := "SELECT id, user_id, number, status, uploaded_at, created_at, updated_at FROM orders WHERE number = :number"

	if err := o.findWithArgs(ctx, tx, args, query, &order); err != nil {
		o.log.Debug("find by number: find with args fail: %w", zap.Error(err))
		return nil
	}

	if order.ID == 0 {
		return nil
	}

	return &order
}

func (o *Order) Create(tx *sqlx.Tx, userID int, status model.OrderStatus, number string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "INSERT INTO orders(user_id, number, status) VALUES(:user_id, :number, :status)"
	args := map[string]interface{}{
		"user_id": userID,
		"number":  number,
		"status":  status,
	}

	if err := o.createWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("order create fail: %w", err)
	}

	return nil
}
