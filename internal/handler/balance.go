package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type balance struct {
	log *zap.Logger
}

func NewBalance(log *zap.Logger) *balance {
	return &balance{log: log}
}

func (b *balance) Get(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Get balance handler called")
}

func (b *balance) Withdraw(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Withdraw balance handler called")
}

func (b *balance) Withdrawals(w http.ResponseWriter, r *http.Request) {
	b.log.Info("Withdrawals balance handler called")
}
