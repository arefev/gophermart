package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service"
	"github.com/arefev/gophermart/internal/service/jwt"
	"go.uber.org/zap"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			m.app.Log.Debug("header Authorization not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		values := strings.Split(header, " ")
		if len(values) != 2 || values[0] != "Bearer" {
			m.app.Log.Debug("get token from header fail")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		login, err := jwt.NewToken(m.app.Conf.TokenSecret).Parse(values[1]).GetLogin()
		if err != nil {
			m.app.Log.Debug("get login fail", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := m.getUser(r.Context(), login)
		if err != nil {
			m.app.Log.Debug("get user fail", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), model.User{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) getUser(ctx context.Context, login string) (*model.User, error) {
	var user *model.User

	user, err := service.NewAuth(m.app).GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get user fail: %w", err)
	}

	return user, nil
}
