package response

import (
	"time"

	"github.com/arefev/gophermart/internal/model"
)

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func NewOrder(o *model.Order) Order {
	return Order{
        Number:     o.Number,
        Status:     o.Status.String(),
        UploadedAt: o.UploadedAt,
    }
}

func NewOrders(l []model.Order) *[]Order {
	orders := make([]Order, 0, len(l))
    for _, o := range l {
        orders = append(orders, NewOrder(&o))
    }
    return &orders
}