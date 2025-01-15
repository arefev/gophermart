package handler

import (
	"errors"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/service"
	"go.uber.org/zap"
)

type user struct {
	app *application.App
}

func NewUser(app *application.App) *user {
	return &user{
		app: app,
	}
}

func (u *user) Register(w http.ResponseWriter, r *http.Request) {
	user, err := service.NewRegister(u.app).FromRequest(r)

	switch {
	case errors.Is(err, service.ErrRegisterUserExists):
		w.WriteHeader(http.StatusConflict)
		return
	case errors.Is(err, service.ErrRegisterJSONDecodeFail), errors.Is(err, service.ErrRegisterValidateFail):
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		u.app.Log.Error("Register user handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := service.NewAuth(u.app).Authorize(r.Context(), user.Login, user.Password)
	if err != nil {
		u.app.Log.Error("Register user handler authorize", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token.AccessToken)
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	token, err := service.NewAuth(u.app).FromRequest(r)

	switch {
	case errors.Is(err, service.ErrAuthUserNotFound):
		w.WriteHeader(http.StatusUnauthorized)
		return
	case errors.Is(err, service.ErrAuthJSONDecodeFail), errors.Is(err, service.ErrAuthValidateFail):
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		u.app.Log.Error("Login user handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token.AccessToken)
}
