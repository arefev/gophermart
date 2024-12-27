package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			m.Log.Debug("header Authorization not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		values := strings.Split(header, " ")
		if len(values) != 2 || values[0] != "Bearer" {
			m.Log.Debug("get token from header fail")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := m.getTokenClaims(values[1])
		if err != nil {
			m.Log.Debug("get token claims fail", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		login, err := m.getLogin(*claims)
		if err != nil {
			m.Log.Debug("get login fail", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := m.getUser(login)
		if err != nil {
			m.Log.Debug("get user fail", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), model.User{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) getTokenClaims(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(m.Conf.TokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parse fail: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("token claims not found")
	}

	return &claims, nil
}

func (m *Middleware) getLogin(claims jwt.MapClaims) (string, error) {
	errLoginNotFound := errors.New("login not found")
	value, ok := claims["login"]
	if !ok {
		return "", errLoginNotFound
	}

	login, ok := value.(string)
	if !ok {
		return "", errLoginNotFound
	}

	return login, nil
}

func (m *Middleware) getUser(login string) (*model.User, error) {
	var user *model.User
	rep := repository.NewUser(m.Log)

	err := db.Transaction(func(tx *sqlx.Tx) error {
		user = rep.FindByLogin(tx, login)
		if user == nil {
			return errors.New("user not found")
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("db transaction fail: %w", err)
	}

	return user, nil
}
