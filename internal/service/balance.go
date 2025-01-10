package service

import (
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
)

type UserBalanceFinder interface {
	FindBalanceByUserID(tx *sqlx.Tx, userID int) *model.Balance
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

	var balance *model.Balance
	err = db.Transaction(func(tx *sqlx.Tx) error {
		balance = ub.Rep.FindBalanceByUserID(tx, user.ID)
		if balance == nil {
			return fmt.Errorf("user balance not found")
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction fail: %w", err)
	}

	return balance, nil
}
