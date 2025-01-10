package handler

import (
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
	rep := repository.NewUser(b.log)
	s := service.NewUserBalance(rep)

	balance, err := s.FromRequest(r)
	if err != nil {
		b.log.Error("Find balance handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	service.JSONResponse(w, balance)

	b.log.Info("Get balance handler called")
}

func (b *balance) Withdraw(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Withdraw balance handler called")
}

func (b *balance) Withdrawals(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Withdrawals balance handler called")
}
