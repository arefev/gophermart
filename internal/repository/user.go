package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type User struct {
	log *zap.Logger
	*Base
}

func NewUser(log *zap.Logger) *User {
	return &User{
		log:  log,
		Base: NewBase(log),
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
	arg := map[string]interface{}{"login": login}

	if err := u.findWithArgs(ctx, tx, arg, query, &user); err != nil {
		u.log.Debug("find by login: find with args fail: %w", zap.Error(err))
		return nil
	}

	if user.ID == 0 {
		return nil
	}

	return &user
}

func (u *User) Create(tx *sqlx.Tx, login, password string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := "INSERT INTO users(login, password) VALUES(:login, :password)"
	args := map[string]interface{}{
		"login":    login,
		"password": password,
	}

	res, err := u.createWithArgs(ctx, tx, args, query)
	if err != nil {
		return 0, fmt.Errorf("user create fail: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("user create fail: %w", err)
	}

	return id, nil
}
