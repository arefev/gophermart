package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"go.uber.org/zap"
)

type User struct {
	log *zap.Logger
	*Base
}

func NewUser(tr TxGetter, log *zap.Logger) *User {
	return &User{
		log:  log,
		Base: NewBase(tr, log),
	}
}

func (u *User) Exists(ctx context.Context, login string) bool {
	_, ok := u.FindByLogin(ctx, login)
	return ok
}

func (u *User) FindByLogin(ctx context.Context, login string) (*model.User, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	user := model.User{}
	query := "SELECT id, login, password, created_at, updated_at FROM users WHERE login = :login"
	arg := map[string]interface{}{"login": login}

	ok, err := u.findWithArgs(ctx, arg, query, &user)
	if err != nil {
		u.log.Debug("find by login: find with args fail: %w", zap.Error(err))
		return nil, false
	}

	return &user, ok
}

func (u *User) Create(ctx context.Context, login, password string) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
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

	if err := u.execWithArgs(ctx, args, query); err != nil {
		return fmt.Errorf("user create fail: %w", err)
	}

	return nil
}
