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

func (o *Order) FindByNumber(tx *sqlx.Tx, number string) (*model.Order, bool) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	order := model.Order{}
	args := map[string]any{"number": number}
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at, created_at, updated_at 
		FROM orders 
		WHERE number = :number
	`

	ok, err := o.findWithArgs(ctx, tx, args, query, &order)
	if err != nil {
		o.log.Debug("find by number: find with args fail: %w", zap.Error(err))
		return nil, false
	}

	return &order, ok
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

	if err := o.execWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("order create fail: %w", err)
	}

	return nil
}

func (o *Order) List(tx *sqlx.Tx, userID int) []model.Order {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	var list []model.Order
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at, created_at, updated_at 
		FROM orders 
		WHERE user_id = :user_id 
		ORDER BY uploaded_at DESC
	`
	args := map[string]interface{}{
		"user_id": userID,
	}

	if err := o.getWithArgs(ctx, tx, args, query, &list); err != nil {
		o.log.Debug("order list fail: get with args fail", zap.Error(err))
		return []model.Order{}
	}

	return list
}

func (o *Order) WithStatusNew(tx *sqlx.Tx) []model.Order {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	var list []model.Order
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at, created_at, updated_at 
		FROM orders 
		WHERE status = :status 
		ORDER BY uploaded_at DESC
	`
	args := map[string]interface{}{
		"status": model.OrderStatusNew,
	}

	if err := o.getWithArgs(ctx, tx, args, query, &list); err != nil {
		o.log.Debug("with status new fail: get with args fail", zap.Error(err))
		return []model.Order{}
	}

	return list
}

func (o *Order) AccrualByID(tx *sqlx.Tx, sum float64, status model.OrderStatus, id int) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "UPDATE orders SET accrual = :accrual, status = :status, updated_at = CURRENT_TIMESTAMP WHERE id = :id"
	args := map[string]interface{}{
		"accrual": sum,
		"id":      id,
		"status":  status,
	}

	if err := o.execWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("accrual by id fail: %w", err)
	}

	return nil
}

func (o *Order) CreateWithdrawal(tx *sqlx.Tx, userID int, number string, sum float64) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "INSERT INTO withdrawals(user_id, number, sum) VALUES(:user_id, :number, :sum)"
	args := map[string]interface{}{
		"user_id": userID,
		"number":  number,
		"sum":     sum,
	}

	if err := o.execWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("create fail: %w", err)
	}

	return nil
}

func (o *Order) GetWithdrawalsByUserID(tx *sqlx.Tx, userID int) []model.Withdrawal {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	var list []model.Withdrawal
	query := `
		SELECT 
			id,
			user_id,
			sum, 
			processed_at,
			created_at,
			updated_at,
			number
		FROM withdrawals
		WHERE user_id = :user_id
	`
	args := map[string]interface{}{
		"user_id": userID,
	}

	if err := o.getWithArgs(ctx, tx, args, query, &list); err != nil {
		o.log.Debug("get withdrawals by user id fail: get with args fail", zap.Error(err))
		return []model.Withdrawal{}
	}

	return list
}
