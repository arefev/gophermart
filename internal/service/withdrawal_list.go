package service

import (
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
)

type WithdrawalGetter interface {
	GetWithdrawalsByUserID(tx *sqlx.Tx, userID int) []model.Withdrawal
}

type WithdrawalList struct {
	rep  WithdrawalGetter
	list []model.Order
}

func NewWithdrawalList(rep WithdrawalGetter) *WithdrawalList {
	return &WithdrawalList{
		rep:  rep,
		list: make([]model.Order, 0),
	}
}

func (wl *WithdrawalList) FromRequest(r *http.Request) ([]model.Withdrawal, error) {
	user, err := UserWithContext(r.Context())
	if err != nil {
		return []model.Withdrawal{}, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	}

	var list []model.Withdrawal
	err = db.Transaction(func(tx *sqlx.Tx) error {
		list = wl.rep.GetWithdrawalsByUserID(tx, user.ID)

		return nil
	})

	if err != nil {
		return []model.Withdrawal{}, fmt.Errorf("transaction fail: %w", err)
	}

	return list, nil
}
