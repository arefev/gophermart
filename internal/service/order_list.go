package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/helper"
	"github.com/arefev/gophermart/internal/model"
)

type OrderList struct {
	app  *application.App
	list []model.Order
}

func NewOrderList(app *application.App) *OrderList {
	return &OrderList{
		app:  app,
		list: make([]model.Order, 0),
	}
}

func (s *OrderList) FromRequest(r *http.Request) ([]model.Order, error) {
	user, err := helper.UserWithContext(r.Context())
	if err != nil {
		return []model.Order{}, helper.ErrUserNotFound
	}

	var orders []model.Order
	err = s.app.TrManager.Do(r.Context(), func(ctx context.Context) error {
		orders = s.app.Rep.Order.List(ctx, user.ID)

		return nil
	})

	if err != nil {
		return []model.Order{}, fmt.Errorf("order list transaction fail: %w", err)
	}

	return orders, nil
}
