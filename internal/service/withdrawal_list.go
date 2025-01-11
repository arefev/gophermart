package service

import (
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
)

type WithdrawalGetter interface {
	GetWithdrawalsByUserID(tx *sqlx.Tx, userID int) []model.WithdrawalWithOrderNumber
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

func (wl *WithdrawalList) FromRequest(r *http.Request) ([]model.WithdrawalWithOrderNumber, error) {
	user, err := UserWithContext(r.Context())
	if err != nil {
		return []model.WithdrawalWithOrderNumber{}, fmt.Errorf("user not found in context: %w", err)
	}

	var list []model.WithdrawalWithOrderNumber
	err = db.Transaction(func(tx *sqlx.Tx) error {
		list = wl.rep.GetWithdrawalsByUserID(tx, user.ID)

		return nil
	})

	if err != nil {
		return []model.WithdrawalWithOrderNumber{}, fmt.Errorf("transaction fail: %w", err)
	}

	return list, nil
}
