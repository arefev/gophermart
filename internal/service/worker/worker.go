package worker

import (
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
	WithStatusNew(tx *sqlx.Tx) []model.Order
	AccrualByID(tx *sqlx.Tx, sum float64, status model.OrderStatus, id int) error
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type worker struct {
	rep OrderNewFinder
	log *zap.Logger
}

func NewWorker(log *zap.Logger, rep OrderNewFinder) *worker {
	return &worker{
		rep: rep,
		log: log,
	}
}

func (w *worker) Run() error {
	readTime := time.NewTicker(time.Duration(2) * time.Second).C

	for range readTime {
		fmt.Println("read time")
		orders := w.getNewOrders()
		w.checkOrders(*orders)
	}

	return nil
}

func (w *worker) getNewOrders() *[]model.Order {
	var orders []model.Order
	err := db.Transaction(func(tx *sqlx.Tx) error {
		orders = w.rep.WithStatusNew(tx)
		return nil
	})

	if err != nil {
		w.log.Error("getNewOrders transaction fail: %w", zap.Error(err))
		return nil
	}

	return &orders
}

func (w *worker) checkOrders(orders []model.Order) {
	for _, order := range orders {
		response, err := w.getStatus(order.Number)
		if err != nil {
			w.log.Error("check order status fail", zap.Error(err))
			continue
		}

		if err := w.updateOrder(order.ID, response); err != nil {
			w.log.Error("update order fail", zap.Error(err))
            continue
        }
	}
}

func (w *worker) getStatus(number string) (*OrderResponse, error) {
	res := OrderResponse{}
	client := resty.New()
	response, err := client.R().
		SetResult(&res).
		Get("http://localhost:8082/api/orders/" + number)

	if err != nil {
		return nil, fmt.Errorf("check order status fail: %w", err)
	}

	if response.StatusCode() != http.StatusOK {
		res.Status = model.OrderStatusInvalid.String()
	}

	return &res, nil
}

func (w *worker) updateOrder(id int, fields *OrderResponse) error {
	err := db.Transaction(func(tx *sqlx.Tx) error {
		status := model.OrderStatusFromString(fields.Status)
		if err := w.rep.AccrualByID(tx, fields.Accrual, status, id); err != nil {
			return fmt.Errorf("update order accrual fail: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("update order transaction fail: %w", err)
	}

	return nil
}
