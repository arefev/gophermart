package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type OrderNewFinder interface {
	WithStatusNew(ctx context.Context, tx *sqlx.Tx) []model.Order
	AccrualByID(ctx context.Context, tx *sqlx.Tx, sum float64, status model.OrderStatus, id int) error
}

type UserBalanceFinder interface {
	FindByUserID(ctx context.Context, tx *sqlx.Tx, userID int) (*model.Balance, bool)
	UpdateByID(ctx context.Context, tx *sqlx.Tx, id int, current, withdrawn float64) error
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type worker struct {
	orderRep       OrderNewFinder
	balanceRep     UserBalanceFinder
	log            *zap.Logger
	accrualAddress string
	pollInterval   int
}

func NewWorker(log *zap.Logger, pollInterval int, accrualAddress string,
	orderRep OrderNewFinder, balanceRep UserBalanceFinder) *worker {
	return &worker{
		orderRep:       orderRep,
		balanceRep:     balanceRep,
		log:            log,
		pollInterval:   pollInterval,
		accrualAddress: accrualAddress,
	}
}

func (w *worker) Run(ctx context.Context) error {
	w.log.Info("Worker started")

	pollTime := time.NewTicker(time.Duration(w.pollInterval) * time.Second).C

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Worker stopped")
			return fmt.Errorf("worker stopped: %w", ctx.Err())
		case <-pollTime:
			w.checkOrders(ctx, *w.getNewOrders(ctx))
		}
	}
}

func (w *worker) getNewOrders(ctx context.Context) *[]model.Order {
	var orders []model.Order
	err := db.Transaction(func(tx *sqlx.Tx) error {
		orders = w.orderRep.WithStatusNew(ctx, tx)
		return nil
	})

	if err != nil {
		w.log.Error("getNewOrders transaction fail: %w", zap.Error(err))
		return nil
	}

	return &orders
}

func (w *worker) checkOrders(ctx context.Context, orders []model.Order) {
	for i := range orders {
		response, err := w.getStatus(orders[i].Number)
		if err != nil {
			w.log.Error("check order status fail", zap.Error(err))
			continue
		}

		if err := w.accrual(ctx, &orders[i], response); err != nil {
			w.log.Error("update order fail", zap.Error(err))
			continue
		}
	}
}

func (w *worker) getStatus(number string) (*OrderResponse, error) {
	url := "http://" + w.accrualAddress + "/api/orders/" + number
	res := OrderResponse{}
	client := resty.New()
	response, err := client.R().
		SetResult(&res).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("check order status fail: %w", err)
	}

	if response.StatusCode() != http.StatusOK {
		res.Status = model.OrderStatusInvalid.String()
	}

	return &res, nil
}

func (w *worker) accrual(ctx context.Context, order *model.Order, fields *OrderResponse) error {
	status := model.OrderStatusFromString(fields.Status)
	if status != model.OrderStatusProcessed && status != model.OrderStatusInvalid {
		return nil
	}

	err := db.Transaction(func(tx *sqlx.Tx) error {
		if status == model.OrderStatusProcessed {
			balance, ok := w.balanceRep.FindByUserID(ctx, tx, order.UserID)
			if !ok {
				return errors.New("user balance not found")
			}

			current := fields.Accrual + balance.Current
			if err := w.balanceRep.UpdateByID(ctx, tx, balance.ID, current, balance.Withdrawn); err != nil {
				return fmt.Errorf("update user balance fail: %w", err)
			}
		}

		if err := w.orderRep.AccrualByID(ctx, tx, fields.Accrual, status, order.ID); err != nil {
			return fmt.Errorf("update order accrual fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("update order transaction fail: %w", err)
	}

	return nil
}
