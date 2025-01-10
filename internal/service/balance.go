package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
)

type UserBalanceFinder interface {
	FindByUserID(tx *sqlx.Tx, userID int) (*model.Balance, bool)
}

type UserBalance struct {
	Rep UserBalanceFinder
}

func NewUserBalance(rep UserBalanceFinder) *UserBalance {
	return &UserBalance{
		Rep: rep,
	}
}

func (ub *UserBalance) FromRequest(req *http.Request) (*model.Balance, error) {
	user, err := UserWithContext(req.Context())
	if err != nil {
		return nil, fmt.Errorf("user not found in context: %w", err)
	}

	balance, err := ub.FindByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("find balance from request fail: %w", err)
	}

	return balance, nil
}

func (ub *UserBalance) FindByUserID(userID int) (*model.Balance, error) {
	var balance *model.Balance
	var ok bool
	err := db.Transaction(func(tx *sqlx.Tx) error {
		balance, ok = ub.Rep.FindByUserID(tx, userID)
		if !ok {
			return errors.New("user balance not found")
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction fail: %w", err)
	}

	return balance, nil
}
