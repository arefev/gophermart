package service

import (
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
)

type OrderGetter interface {
	List(tx *sqlx.Tx, userID int) []model.Order
}

type OrderList struct {
	rep  OrderGetter
	list []model.Order
}

func NewOrderList(rep OrderGetter) *OrderList {
	return &OrderList{
		rep:  rep,
		list: make([]model.Order, 0),
	}
}

func (s *OrderList) FromRequest(r *http.Request) ([]model.Order, error) {
	user, err := UserWithContext(r.Context())
	if err != nil {
		return []model.Order{}, fmt.Errorf("user not found in context: %w", err)
	}

	var orders []model.Order
	err = db.Transaction(func(tx *sqlx.Tx) error {
		orders = s.rep.List(tx, user.ID)

		return nil
	})

	if err != nil {
		return []model.Order{}, fmt.Errorf("transaction fail: %w", err)
	}

	return orders, nil
}
