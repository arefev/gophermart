package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotEnoughBalance     = errors.New("not enough balance")
	ErrOrderNotFound        = errors.New("order not found")
	ErrValidationWithdrawal = errors.New("validation withdrawal fail")
)

type UserBalanceFinder interface {
	FindByUserID(tx *sqlx.Tx, userID int) (*model.Balance, bool)
	UpdateByID(tx *sqlx.Tx, id int, current, withdrawn float64) error
}

type OrderFinder interface {
	FindByNumber(tx *sqlx.Tx, number string) (*model.Order, bool)
	CreateWithdrawal(tx *sqlx.Tx, orderID int, sum float64) error
}

type WithdrawalRequest struct {
	Number string  `json:"number" validate:"required,alphanum,gte=3,lte=50"`
	Sum    float64 `json:"sum" validate:"required"`
}

type UserBalance struct {
	Rep      UserBalanceFinder
	OrderRep OrderFinder
}

func NewUserBalance(rep UserBalanceFinder) *UserBalance {
	return &UserBalance{
		Rep: rep,
	}
}

func (ub *UserBalance) SetOrderRep(rep OrderFinder) *UserBalance {
	ub.OrderRep = rep
	return ub
}

func (ub *UserBalance) FromRequest(req *http.Request) (*model.Balance, error) {
	user, err := UserWithContext(req.Context())
	if err != nil {
		return nil, ErrUserNotFound
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
		return nil, fmt.Errorf("find by user id %w: %w", db.ErrTransactionFail, err)
	}

	return balance, nil
}

func (ub *UserBalance) WithdrawalFromRequest(req *http.Request) error {
	wr, err := ub.validateWithdrawal(req)
	if err != nil {
		return fmt.Errorf("validate withdrawal from request fail: %w", err)
	}

	user, err := UserWithContext(req.Context())
	if err != nil {
		return ErrUserNotFound
	}

	if err := ub.Withdrawal(user, wr); err != nil {
		return fmt.Errorf("withdrawal from request fail: %w", err)
	}

	return nil
}

func (ub *UserBalance) Withdrawal(user *model.User, wr *WithdrawalRequest) error {
	balance, err := ub.FindByUserID(user.ID)
	if err != nil {
		return fmt.Errorf("balance not found: %w", err)
	}

	err = db.Transaction(func(tx *sqlx.Tx) error {
		order, ok := ub.OrderRep.FindByNumber(tx, wr.Number)
		if !ok || order.UserID != user.ID {
			return ErrOrderNotFound
		}

		if balance.Current < wr.Sum {
			return ErrNotEnoughBalance
		}

		current := balance.Current - wr.Sum
		withdrawn := balance.Withdrawn + wr.Sum
		if err := ub.Rep.UpdateByID(tx, balance.ID, current, withdrawn); err != nil {
			return fmt.Errorf("balance update fail: %w", err)
		}

		if err := ub.OrderRep.CreateWithdrawal(tx, order.ID, wr.Sum); err != nil {
			return fmt.Errorf("create withdrawal fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("withdrawal %w: %w", db.ErrTransactionFail, err)
	}

	return nil
}

func (ub *UserBalance) validateWithdrawal(r *http.Request) (*WithdrawalRequest, error) {
	rOrder := WithdrawalRequest{}
	d := json.NewDecoder(r.Body)

	if err := d.Decode(&rOrder); err != nil {
		return nil, fmt.Errorf("decode json body fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(rOrder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidationWithdrawal, err)
	}

	return &rOrder, nil
}
