package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"go.uber.org/zap"
)

type Order struct {
	log *zap.Logger
	*Base
}

func NewOrder(tr TxGetter, log *zap.Logger) *Order {
	return &Order{
		log:  log,
		Base: NewBase(tr, log),
	}
}

func (o *Order) FindByNumber(ctx context.Context, number string) (*model.Order, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	order := model.Order{}
	args := map[string]any{"number": number}
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at, created_at, updated_at 
		FROM orders 
		WHERE number = :number
	`

	ok, err := o.findWithArgs(ctx, args, query, &order)
	if err != nil {
		o.log.Debug("find by number: find with args fail: %w", zap.Error(err))
		return nil, false
	}

	return &order, ok
}

func (o *Order) Create(ctx context.Context, userID int, status model.OrderStatus, number string) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	query := "INSERT INTO orders(user_id, number, status) VALUES(:user_id, :number, :status)"
	args := map[string]interface{}{
		"user_id": userID,
		"number":  number,
		"status":  status,
	}

	if err := o.execWithArgs(ctx, args, query); err != nil {
		return fmt.Errorf("order create fail: %w", err)
	}

	return nil
}

func (o *Order) List(ctx context.Context, userID int) []model.Order {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
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

	if err := o.getWithArgs(ctx, args, query, &list); err != nil {
		o.log.Debug("order list fail: get with args fail", zap.Error(err))
		return []model.Order{}
	}

	return list
}

func (o *Order) WithStatusNew(ctx context.Context) []model.Order {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
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

	if err := o.getWithArgs(ctx, args, query, &list); err != nil {
		o.log.Debug("with status new fail: get with args fail", zap.Error(err))
		return []model.Order{}
	}

	return list
}

func (o *Order) AccrualByID(ctx context.Context, sum float64, status model.OrderStatus, id int) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	query := "UPDATE orders SET accrual = :accrual, status = :status, updated_at = CURRENT_TIMESTAMP WHERE id = :id"
	args := map[string]interface{}{
		"accrual": sum,
		"id":      id,
		"status":  status,
	}

	if err := o.execWithArgs(ctx, args, query); err != nil {
		return fmt.Errorf("accrual by id fail: %w", err)
	}

	return nil
}

func (o *Order) CreateWithdrawal(ctx context.Context, userID int, number string, sum float64) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	query := "INSERT INTO withdrawals(user_id, number, sum) VALUES(:user_id, :number, :sum)"
	args := map[string]interface{}{
		"user_id": userID,
		"number":  number,
		"sum":     sum,
	}

	if err := o.execWithArgs(ctx, args, query); err != nil {
		return fmt.Errorf("create fail: %w", err)
	}

	return nil
}

func (o *Order) GetWithdrawalsByUserID(ctx context.Context, userID int) []model.Withdrawal {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
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

	if err := o.getWithArgs(ctx, args, query, &list); err != nil {
		o.log.Debug("get withdrawals by user id fail: get with args fail", zap.Error(err))
		return []model.Withdrawal{}
	}

	return list
}
