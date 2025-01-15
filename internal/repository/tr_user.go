package repository

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TxGetter interface {
	FromCtx(context.Context) (*sqlx.Tx, error)
}

type TrUser struct {
	log *zap.Logger
	tr  TxGetter
	*Base
}

func NewTrUser(tr TxGetter, log *zap.Logger) *TrUser {
	return &TrUser{
		log:  log,
		tr:   tr,
		Base: NewBase(log),
	}
}

func (utr *TrUser) FindByLogin(ctx context.Context, login string) (*model.User, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	user := model.User{}
	query := "SELECT id, login, password, created_at, updated_at FROM users WHERE login = :login"
	arg := map[string]any{"login": login}

	tx, _ := utr.tr.FromCtx(ctx)

	ok, err := utr.findWithArgs(ctx, tx, arg, query, &user)
	if err != nil {
		utr.log.Debug("find by login: find with args fail: %w", zap.Error(err))
		return nil, false
	}

	return &user, ok
}

func (utr *TrUser) Create(ctx context.Context, login, password string) error {
	ctx, cancel := context.WithTimeout(ctx, timeCancel)
	defer cancel()

	query := `
		INSERT INTO users(login, password) VALUES(:login, :password)
	`
	args := map[string]interface{}{
		"login":    login,
		"password": password,
	}

	tx, _ := utr.tr.FromCtx(ctx)

	if err := utr.execWithArgs(ctx, tx, args, query); err != nil {
		return fmt.Errorf("user create fail: %w", err)
	}

	return nil
}
