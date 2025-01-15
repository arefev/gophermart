package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/helper"
	"github.com/arefev/gophermart/internal/model"
)

type WithdrawalList struct {
	app  *application.App
	list []model.Order
}

func NewWithdrawalList(app *application.App) *WithdrawalList {
	return &WithdrawalList{
		app:  app,
		list: make([]model.Order, 0),
	}
}

func (wl *WithdrawalList) FromRequest(r *http.Request) ([]model.Withdrawal, error) {
	user, err := helper.UserWithContext(r.Context())
	if err != nil {
		return []model.Withdrawal{}, helper.ErrUserNotFound
	}

	var list []model.Withdrawal
	err = wl.app.TrManager.Do(r.Context(), func(ctx context.Context) error {
		list = wl.app.Rep.Order.GetWithdrawalsByUserID(ctx, user.ID)

		return nil
	})

	if err != nil {
		return []model.Withdrawal{}, fmt.Errorf("withdrawal list transaction fail: %w", err)
	}

	return list, nil
}
