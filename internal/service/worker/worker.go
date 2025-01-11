package worker

import (
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
	WithStatusNew(tx *sqlx.Tx) []model.Order
	AccrualByID(tx *sqlx.Tx, sum float64, status model.OrderStatus, id int) error
}

type UserBalanceFinder interface {
	FindByUserID(tx *sqlx.Tx, userID int) (*model.Balance, bool)
	UpdateByID(tx *sqlx.Tx, id int, current, withdrawn float64) error
}

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type worker struct {
	orderRep   OrderNewFinder
	balanceRep UserBalanceFinder
	log        *zap.Logger
}

func NewWorker(log *zap.Logger, orderRep OrderNewFinder, balanceRep UserBalanceFinder) *worker {
	return &worker{
		orderRep:   orderRep,
		balanceRep: balanceRep,
		log:        log,
	}
}

func (w *worker) Run() {
	const interval = 2
	checkTime := time.NewTicker(time.Duration(interval) * time.Second).C

	for range checkTime {
		fmt.Println("check time")
		orders := w.getNewOrders()
		w.checkOrders(*orders)
	}
}

func (w *worker) getNewOrders() *[]model.Order {
	var orders []model.Order
	err := db.Transaction(func(tx *sqlx.Tx) error {
		orders = w.orderRep.WithStatusNew(tx)
		return nil
	})

	if err != nil {
		w.log.Error("getNewOrders transaction fail: %w", zap.Error(err))
		return nil
	}

	return &orders
}

func (w *worker) checkOrders(orders []model.Order) {
	for i := range orders {
		response, err := w.getStatus(orders[i].Number)
		if err != nil {
			w.log.Error("check order status fail", zap.Error(err))
			continue
		}

		if err := w.accrual(&orders[i], response); err != nil {
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

func (w *worker) accrual(order *model.Order, fields *OrderResponse) error {
	status := model.OrderStatusFromString(fields.Status)
	if status != model.OrderStatusProcessed && status != model.OrderStatusInvalid {
		return nil
	}

	err := db.Transaction(func(tx *sqlx.Tx) error {
		if status == model.OrderStatusProcessed {
			balance, ok := w.balanceRep.FindByUserID(tx, order.UserID)
			if !ok {
				return errors.New("user balance not found")
			}

			current := fields.Accrual + balance.Current
			if err := w.balanceRep.UpdateByID(tx, balance.ID, current, balance.Withdrawn); err != nil {
				return fmt.Errorf("update user balance fail: %w", err)
			}
		}

		if err := w.orderRep.AccrualByID(tx, fields.Accrual, status, order.ID); err != nil {
			return fmt.Errorf("update order accrual fail: %w", err)
		}
		
		return nil
	})

	if err != nil {
		return fmt.Errorf("update order transaction fail: %w", err)
	}

	return nil
}
