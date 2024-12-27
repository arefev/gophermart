package handler

import (
	"errors"
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
	err := service.NewRegister(rep, u.log).FromRequest(r)

	switch {
	case errors.Is(err, service.ErrRegisterUserExists):
		w.WriteHeader(http.StatusConflict)
		return
	case errors.Is(err, service.ErrRegisterJsonDecodeFail), errors.Is(err, service.ErrRegisterValidateFail):
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		u.log.Error("Register user handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	u.log.Info("Login user handler called")

	rep := repository.NewUser(u.log)
	token, err := service.NewAuth(rep, u.log).FromRequest(r)

	switch {
	case errors.Is(err, service.ErrAuthUserNotFound):
		w.WriteHeader(http.StatusUnauthorized)
		return
	case errors.Is(err, service.ErrAuthJsonDecodeFail), errors.Is(err, service.ErrAuthValidateFail):
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		u.log.Error("Login user handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer " + token)
}
