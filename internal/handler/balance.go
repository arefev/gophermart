package handler

import (
	"errors"
	"net/http"

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
	balance, err := service.NewUserBalance(b.app).FromRequest(r)

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
	err := service.NewUserBalance(b.app).WithdrawalFromRequest(r)

	switch {
	case errors.Is(err, service.ErrNotEnoughBalance):
		w.WriteHeader(http.StatusPaymentRequired)
		return
	case errors.Is(err, service.ErrValidationWithdrawal):
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case err != nil:
		b.app.Log.Error("Withdraw balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (b *balance) Withdrawals(w http.ResponseWriter, r *http.Request) {
	list, err := service.NewWithdrawalList(b.app).FromRequest(r)

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
