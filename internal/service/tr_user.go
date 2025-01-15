package service

import (
	"context"
	"fmt"

	"github.com/arefev/gophermart/internal/trm"
)

type Transaction interface {
	Do(ctx context.Context, action trm.TrAction) error
}

type UserRepo interface {
	Create(ctx context.Context, login, password string) error
}

type UserCreateAction struct {
	tr      Transaction
	userRep UserRepo
}

func NewUserCreateAction(tr Transaction, userRep UserRepo) *UserCreateAction {
	return &UserCreateAction{
		tr:      tr,
		userRep: userRep,
	}
}

func (uca *UserCreateAction) Run(ctx context.Context, login, pwd string) error {
	err := uca.tr.Do(ctx, func(ctx context.Context) error {
		return uca.userRep.Create(ctx, login, pwd)
	})

	if err != nil {
		return fmt.Errorf("user create action run fail: %w", err)
	}

	return nil
}
