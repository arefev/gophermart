package handler

import (
	"net/http"

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
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	u.log.Info("Login user handler called")
}
