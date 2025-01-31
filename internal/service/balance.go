package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/trm"
)

type balanceService struct {
	app *application.App
}

func NewBalanceService(app *application.App) *balanceService {
	return &balanceService{
		app: app,
	}
}

func (bs *balanceService) FindByUserID(ctx context.Context, userID int) (*model.Balance, error) {
	var balance *model.Balance
	var ok bool
	err := bs.app.TrManager.Do(ctx, func(ctx context.Context) error {
		balance, ok = bs.app.Rep.Balance.FindByUserID(ctx, userID)
		if !ok {
			return errors.New("user balance not found")
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("find by user id %w: %w", trm.ErrTransactionFail, err)
	}

	return balance, nil
}
