package handler

import (
	"errors"
	"net/http"

	b_action "github.com/arefev/gophermart/internal/action/balance"
	w_action "github.com/arefev/gophermart/internal/action/withdrawal"
	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/response"
	"github.com/arefev/gophermart/internal/service"
	"go.uber.org/zap"
)

type balance struct {
	app *application.App
}

func NewBalance(app *application.App) *balance {
	return &balance{app: app}
}

func (b *balance) Find(w http.ResponseWriter, r *http.Request) {
	balance, err := b_action.NewBalanceAction(b.app).Handle(r)

	if err != nil {
		b.app.Log.Error("Find balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := service.JSONResponse(w, balance); err != nil {
		b.app.Log.Error("Find balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (b *balance) Withdraw(w http.ResponseWriter, r *http.Request) {
	err := w_action.NewCreateAction(b.app).Handle(r)

	switch {
	case errors.Is(err, w_action.ErrNotEnoughBalance):
		w.WriteHeader(http.StatusPaymentRequired)
		return
	case errors.Is(err, w_action.ErrValidationWithdrawal):
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case err != nil:
		b.app.Log.Error("Withdraw balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (b *balance) Withdrawals(w http.ResponseWriter, r *http.Request) {
	list, err := w_action.NewListAction(b.app).Handle(r)

	switch {
	case len(list) == 0:
		w.WriteHeader(http.StatusNoContent)
		return
	case err != nil:
		b.app.Log.Error("Withdrawals balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := service.JSONResponse(w, response.NewWithdrawals(list)); err != nil {
		b.app.Log.Error("Withdrawals balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
