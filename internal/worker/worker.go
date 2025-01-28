package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"go.uber.org/zap"
)

type StatusRequest interface {
	Request(ctx context.Context, number string, res *OrderResponse) error
}

type OrderResponse struct {
	Header     http.Header `json:"-"`
	Order      string      `json:"order"`
	Status     string      `json:"status"`
	Accrual    float64     `json:"accrual"`
	HTTPStatus int         `json:"-"`
}

type worker struct {
	app      *application.App
	request  StatusRequest
	job      chan *model.Order
	ticker   *time.Ticker
	isActive bool
}

func NewWorker(app *application.App, r StatusRequest) *worker {
	return &worker{
		app:     app,
		request: r,
	}
}

func (w *worker) Run(ctx context.Context) error {
	w.app.Log.Info("Worker started")

	w.pool(ctx)
	w.ticker = time.NewTicker(w.tickerTime())
	w.isActive = true

	for {
		select {
		case <-ctx.Done():
			w.app.Log.Info("Worker stopped")
			return fmt.Errorf("worker stopped: %w", ctx.Err())
		case <-w.ticker.C:
			w.app.Log.Info("Worker polling")
			w.handle(ctx)
		}
	}
}

func (w *worker) handle(ctx context.Context) {
	w.checkOrders(w.getNewOrders(ctx))
}

func (w *worker) getNewOrders(ctx context.Context) []model.Order {
	var orders []model.Order
	err := w.app.TrManager.Do(ctx, func(ctx context.Context) error {
		orders = w.app.Rep.Order.WithStatusNew(ctx)
		return nil
	})

	if err != nil {
		w.app.Log.Error("getNewOrders transaction fail", zap.Error(err))
		return []model.Order{}
	}

	return orders
}

func (w *worker) checkOrders(orders []model.Order) {
	for i := range orders {
		w.createJob(&orders[i])
	}
}

func (w *worker) getStatus(ctx context.Context, number string) (*OrderResponse, error) {
	res := OrderResponse{}
	err := w.request.Request(ctx, number, &res)
	if err != nil {
		return &OrderResponse{}, fmt.Errorf("get status fail: %w", err)
	}

	return &res, nil
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

func (w *worker) pool(ctx context.Context) {
	limit := w.app.Conf.RateLimit
	w.job = make(chan *model.Order, limit)

	for range limit {
		go w.listener(ctx)
	}
}

func (w *worker) listener(ctx context.Context) {
	for order := range w.job {
		w.runJob(ctx, order)
	}
}

func (w *worker) createJob(order *model.Order) {
	w.job <- order
}

func (w *worker) runJob(ctx context.Context, order *model.Order) {
	response, err := w.getStatus(ctx, order.Number)

	if w.shouldRestart(response) {
		w.restart(response)
		return
	}

	if err != nil {
		w.app.Log.Error("check order status fail", zap.Error(err))
		return
	}

	if err := w.accrual(ctx, order, response); err != nil {
		w.app.Log.Error("update order fail", zap.Error(err))
		return
	}
}

func (w *worker) shouldRestart(r *OrderResponse) bool {
	return r.HTTPStatus == http.StatusTooManyRequests
}

func (w *worker) restart(r *OrderResponse) {
	const waitTime = time.Duration(60) * time.Second
	var wait time.Duration

	t := r.Header.Get("Retry-After")
	d, err := time.ParseDuration(t + "s")
	if err != nil {
		d = waitTime
	}
	wait = d

	w.restartAfter(wait)
}

func (w *worker) restartAfter(d time.Duration) {
	if !w.isActive {
		return
	}

	w.ticker.Stop()
	time.AfterFunc(d, func() {
		w.ticker.Reset(w.tickerTime())
		w.isActive = true
	})

	w.isActive = false
	w.app.Log.Sugar().Infof("worker wait time is %+v", d)
}

func (w *worker) tickerTime() time.Duration {
	return time.Duration(w.app.Conf.PollInterval) * time.Second
}
