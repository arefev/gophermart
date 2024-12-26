package handler

import (
	"net/http"

	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/service"
	"go.uber.org/zap"
)

type user struct {
	log *zap.Logger
}

func NewUser(log *zap.Logger) *user {
	return &user{log: log}
}

func (u *user) Register(w http.ResponseWriter, r *http.Request) {
	u.log.Info("Register user handler called")

	rep := repository.NewUser(u.log)
	if err := service.NewRegister(rep, u.log).FromRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	u.log.Info("Login user handler called")

	rep := repository.NewUser(u.log)
	if err := service.NewAuth(rep, u.log).FromRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
