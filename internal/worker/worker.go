package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"go.uber.org/zap"
)

type StatusRequest interface {
	Request(ctx context.Context, number string, res *OrderResponse) (time.Duration, error)
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type worker struct {
	nextPollingTime time.Time
	app             *application.App
	request         StatusRequest
}

func NewWorker(app *application.App, r StatusRequest) *worker {
	return &worker{
		app:             app,
		nextPollingTime: time.Now(),
		request:         r,
	}
}

func (w *worker) Run(ctx context.Context) error {
	w.app.Log.Info("Worker started")

	pollTime := time.NewTicker(time.Duration(w.app.Conf.PollInterval) * time.Second).C

	for {
		select {
		case <-ctx.Done():
			w.app.Log.Info("Worker stopped")
			return fmt.Errorf("worker stopped: %w", ctx.Err())
		case <-pollTime:
			if time.Since(w.nextPollingTime) >= 0 {
				w.nextPollingTime = time.Now()
				w.app.Log.Info("Worker polling")
				w.Handle(ctx)
			}
		}
	}
}

func (w *worker) Handle(ctx context.Context) {
	w.checkOrders(ctx, w.getNewOrders(ctx))
}

func (w *worker) getNewOrders(ctx context.Context) []model.Order {
	var orders []model.Order
	err := w.app.TrManager.Do(ctx, func(ctx context.Context) error {
		orders = w.app.Rep.Order.WithStatusNew(ctx)
		return nil
	})

	if err != nil {
		w.app.Log.Error("getNewOrders transaction fail", zap.Error(err))
		return nil
	}

	return orders
}

func (w *worker) checkOrders(ctx context.Context, orders []model.Order) {
	for i := range orders {
		response, wait, err := w.getStatus(ctx, orders[i].Number)
		if wait > 0 {
			w.app.Log.Sugar().Infof("worker polling waiting %v", wait)
			w.nextPollingTime = time.Now().Add(wait)
			return
		}

		if err != nil {
			w.app.Log.Error("check order status fail", zap.Error(err))
			continue
		}

		if err := w.accrual(ctx, &orders[i], response); err != nil {
			w.app.Log.Error("update order fail", zap.Error(err))
			continue
		}
	}
}

func (w *worker) getStatus(ctx context.Context, number string) (*OrderResponse, time.Duration, error) {
	res := OrderResponse{}
	wait, err := w.request.Request(ctx, number, &res)
	if err != nil {
		return nil, 0, fmt.Errorf("get status fail: %w", err)
	}

	return &res, wait, nil
}

func (w *worker) accrual(ctx context.Context, order *model.Order, fields *OrderResponse) error {
	status := model.OrderStatusFromString(fields.Status)
	if status != model.OrderStatusProcessed && status != model.OrderStatusInvalid {
		return nil
	}

	err := w.app.TrManager.Do(ctx, func(ctx context.Context) error {
		if status == model.OrderStatusProcessed {
			balance, ok := w.app.Rep.Balance.FindByUserID(ctx, order.UserID)
			if !ok {
				return errors.New("user balance not found")
			}

			current := fields.Accrual + balance.Current
			if err := w.app.Rep.Balance.UpdateByID(ctx, balance.ID, current, balance.Withdrawn); err != nil {
				return fmt.Errorf("update user balance fail: %w", err)
			}
		}

		if err := w.app.Rep.Order.AccrualByID(ctx, fields.Accrual, status, order.ID); err != nil {
			return fmt.Errorf("update order accrual fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("update order transaction fail: %w", err)
	}

	return nil
}
