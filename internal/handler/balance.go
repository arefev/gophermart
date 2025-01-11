package handler

import (
	"errors"
	"net/http"

	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/service"
	"go.uber.org/zap"
)

type balance struct {
	log *zap.Logger
}

func NewBalance(log *zap.Logger) *balance {
	return &balance{log: log}
}

func (b *balance) Find(w http.ResponseWriter, r *http.Request) {
	rep := repository.NewBalance(b.log)
	s := service.NewUserBalance(rep)

	balance, err := s.FromRequest(r)
	if err != nil {
		b.log.Error("Find balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := service.JSONResponse(w, balance); err != nil {
		b.log.Error("Find balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b.log.Info("Find balance handler called")
}

func (b *balance) Withdraw(w http.ResponseWriter, r *http.Request) {
	bRep := repository.NewBalance(b.log)
	oRep := repository.NewOrder(b.log)
	s := service.NewUserBalance(bRep).SetOrderRep(oRep)

	err := s.WithdrawalFromRequest(r)

	switch {
	case errors.Is(err, service.ErrNotEnoughBalance):
		w.WriteHeader(http.StatusPaymentRequired)
		return
	case errors.Is(err, service.ErrValidationWithdrawal):
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case err != nil:
		b.log.Error("Withdraw balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b.log.Info("Withdraw balance handler called")
}

func (b *balance) Withdrawals(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Withdrawals balance handler called")
}
