package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/helper"
	"github.com/arefev/gophermart/internal/model"
	"github.com/go-playground/validator/v10"
)

var (
	ErrOrderCreateUploadedByCurrentUser = errors.New("order already uploaded by current user")
	ErrOrderCreateUploadedByOtherUser   = errors.New("order already uploaded by other user")
	ErrOrderCreateValidateFail          = errors.New("validate fail")
)

type OrderCreateRequest struct {
	Number string `json:"number" validate:"required,alphanum,gte=3,lte=50"`
}

type OrderCreate struct {
	app *application.App
}

func NewOrderCreate(app *application.App) *OrderCreate {
	return &OrderCreate{
		app: app,
	}
}

func (ocr *OrderCreate) FromRequest(r *http.Request) error {
	const errMsg = "order create from request:"
	rOrder, err := ocr.validate(r)
	if err != nil {
		return fmt.Errorf("%w %w", ErrOrderCreateValidateFail, err)
	}

	user, err := helper.UserWithContext(r.Context())
	if err != nil {
		return helper.ErrUserNotFound
	}

	err = ocr.app.TrManager.Do(r.Context(), func(ctx context.Context) error {
		order, ok := ocr.app.Rep.Order.FindByNumber(ctx, rOrder.Number)
		if ok {
			if order.UserID == user.ID {
				return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByCurrentUser)
			}

			return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByOtherUser)
		}

		if err := ocr.app.Rep.Order.Create(ctx, user.ID, model.OrderStatusNew, rOrder.Number); err != nil {
			return fmt.Errorf("%s create fail: %w", errMsg, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s transaction fail: %w", errMsg, err)
	}

	return nil
}

func (ocr *OrderCreate) validate(r *http.Request) (*OrderCreateRequest, error) {
	number, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read body fail: %w", err)
	}

	rOrder := OrderCreateRequest{
		Number: strings.Trim(string(number), " "),
	}

	if err := helper.CheckLuhn(rOrder.Number); err != nil {
		return nil, fmt.Errorf("luhn check fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(rOrder); err != nil {
		return nil, fmt.Errorf("validation fail: %w", err)
	}

	return &rOrder, nil
}
