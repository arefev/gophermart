package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const timeCancel = 15 * time.Second

type User struct {
	log *zap.Logger
}

func NewUser(log *zap.Logger) *User {
	return &User{
		log: log,
	}
}

func (u *User) Exists(tx *sqlx.Tx, login string) bool {
	return u.FindByLogin(tx, login) != nil
}

func (u *User) FindByLogin(tx *sqlx.Tx, login string) *model.User {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	user := model.User{}
	query := "SELECT id, login, password, created_at, updated_at FROM users WHERE login = :login"

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		u.log.Debug("user find by login fail", zap.Error(err))
		return nil
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			u.log.Warn("user find fail", zap.Error(err))
		}
	}()

	arg := map[string]interface{}{"login": login}
	if err := stmt.GetContext(ctx, &user, arg); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			u.log.Debug("user find by login fail", zap.Error(err))
		}

		return nil
	}

	return &user
}

func (u *User) Create(tx *sqlx.Tx, login string, password string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "INSERT INTO users(login, password) VALUES(:login, :password)"
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("user create fail: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			u.log.Warn("user create fail", zap.Error(err))
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		map[string]interface{}{
			"login":    login,
			"password": password,
		},
	)

	if err != nil {
		return fmt.Errorf("user create fail: %w", err)
	}

	return nil
}
