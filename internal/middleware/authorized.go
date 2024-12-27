package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func (m *Middleware) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")
		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		values := strings.Split(header, " ")
		if len(values) != 2 || values[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		login, err := m.getLoginFromToken(values[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := m.getUser(login)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), model.User{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) getLoginFromToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(m.Conf.TokenSecret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("token claims not found")
	}

	login, ok := claims["login"]
	if !ok {
		return "", fmt.Errorf("login not found")
	}

	return login.(string), nil
}

func (m *Middleware) getUser(login string) (*model.User, error) {
	var user *model.User
	var err error
	rep := repository.NewUser(m.Log)

	err = db.Transaction(func(tx *sqlx.Tx) error {
		user, err = rep.FindByLogin(tx, login)
		if err != nil {
			return fmt.Errorf("get user fail: %w", err)
		}
		return nil
	})

	return user, err
}
