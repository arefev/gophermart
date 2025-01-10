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

func (u *User) Create(tx *sqlx.Tx, login, password string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	query := `
		WITH inserted AS (
			INSERT INTO users(login, password) VALUES(:login, :password) RETURNING id
		)
		INSERT INTO users_balance(user_id) SELECT id FROM inserted
	`
	args := map[string]interface{}{
		"login":    login,
		"password": password,
	}

	if err := u.createWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("user create fail: %w", err)
	}

	return nil
}

func (u *User) FindBalanceByUserID(tx *sqlx.Tx, userID int) *model.Balance {
	ctx, cancel := context.WithTimeout(context.TODO(), timeCancel)
	defer cancel()

	balance := model.Balance{}
	query := "SELECT id, user_id, current, withdrawn, created_at, updated_at FROM users_balance WHERE user_id = :user_id"
	arg := map[string]interface{}{"user_id": userID}

	if err := u.findWithArgs(ctx, tx, arg, query, &balance); err != nil {
		u.log.Debug("find balance by user id: find with args fail: %w", zap.Error(err))
		return nil
	}

	if balance.ID == 0 {
		return nil
	}

	return &balance
}
