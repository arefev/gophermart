package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Order struct {
	log *zap.Logger
}

func NewOrder(log *zap.Logger) *Order {
	return &Order{
		log: log,
	}
}

func (o *Order) FindByNumber(tx *sqlx.Tx, number string) *model.Order {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	order := model.Order{}
	query := "SELECT id, user_id, number, status, uploaded_at, created_at, updated_at FROM orders WHERE number = :number"
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		o.log.Debug("order find by number fail", zap.Error(err))
		return nil
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			o.log.Warn("order find fail", zap.Error(err))
		}
	}()

	arg := map[string]interface{}{"number": number}
	if err := stmt.GetContext(ctx, &order, arg); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			o.log.Debug("order find by user and number fail", zap.Error(err))
		}

		return nil
	}

	return &order
}

func (o *Order) Create(tx *sqlx.Tx, userID int, status model.OrderStatus, number string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "INSERT INTO orders(user_id, number, status) VALUES(:user_id, :number, :status)"
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("order create fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			o.log.Warn("order create fail", zap.Error(err))
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		map[string]interface{}{
			"user_id": userID,
			"number":  number,
			"status":  status,
		},
	)

	if err != nil {
		return fmt.Errorf("order create fail: %w", err)
	}

	return nil
}
