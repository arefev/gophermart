package withdrawal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service"
	"github.com/arefev/gophermart/internal/service/alg"
	"github.com/arefev/gophermart/internal/trm"
	"github.com/go-playground/validator/v10"
)

var (
	ErrNotEnoughBalance     = errors.New("not enough balance")
	ErrOrderNotFound        = errors.New("order not found")
	ErrValidationWithdrawal = errors.New("validation withdrawal fail")
)

type CreateRequest struct {
	Order string  `json:"order" validate:"required,alphanum,gte=3,lte=50"`
	Sum   float64 `json:"sum" validate:"required"`
}

type createAction struct {
	app *application.App
}

func NewCreateAction(app *application.App) *createAction {
	return &createAction{
		app: app,
	}
}

func (c *createAction) Handle(r *http.Request) error {
	wr, err := c.validate(r)
	if err != nil {
		return fmt.Errorf("validate withdrawal from request fail: %w", err)
	}

	user, err := service.NewUserService(c.app).Authorized(r.Context())
	if err != nil {
		return service.ErrUserNotAuthorized
	}

	if err := c.withdrawal(r.Context(), user, wr); err != nil {
		return fmt.Errorf("withdrawal from request fail: %w", err)
	}

	return nil
}

func (c *createAction) withdrawal(ctx context.Context, user *model.User, wr *CreateRequest) error {
	balance, err := service.NewBalanceService(c.app).FindByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("balance not found: %w", err)
	}

	err = c.app.TrManager.Do(ctx, func(ctx context.Context) error {
		if balance.Current < wr.Sum {
			return ErrNotEnoughBalance
		}

		current := balance.Current - wr.Sum
		withdrawn := balance.Withdrawn + wr.Sum
		if err := c.app.Rep.Balance.UpdateByID(ctx, balance.ID, current, withdrawn); err != nil {
			return fmt.Errorf("balance update fail: %w", err)
		}

		if err := c.app.Rep.Order.CreateWithdrawal(ctx, user.ID, wr.Order, wr.Sum); err != nil {
			return fmt.Errorf("create withdrawal fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("withdrawal %w: %w", trm.ErrTransactionFail, err)
	}

	return nil
}

func (c *createAction) validate(r *http.Request) (*CreateRequest, error) {
	rOrder := CreateRequest{}
	d := json.NewDecoder(r.Body)

	if err := d.Decode(&rOrder); err != nil {
		return nil, fmt.Errorf("decode json body fail: %w", err)
	}

	if err := alg.CheckLuhn(rOrder.Order); err != nil {
		return nil, fmt.Errorf("luhn check fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(rOrder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidationWithdrawal, err)
	}

	return &rOrder, nil
}
