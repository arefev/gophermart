package response

import (
	"time"

	"github.com/arefev/gophermart/internal/model"
)

type Order struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
}

func NewOrder(o *model.Order) Order {
	return Order{
		Number:     o.Number,
		Status:     o.Status.String(),
		Accrual:    o.Accrual.Float64,
		UploadedAt: o.UploadedAt,
	}
}

func NewOrders(l []model.Order) *[]Order {
	orders := make([]Order, 0, len(l))
	for i := range l {
		orders = append(orders, NewOrder(&l[i]))
	}
	return &orders
}
