package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arefev/gophermart/internal/helper"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var (
	ErrOrderCreateUploadedByCurrentUser = errors.New("order already uploaded by current user")
	ErrOrderCreateUploadedByOtherUser   = errors.New("order already uploaded by other user")
	ErrOrderCreateValidateFail          = errors.New("validate fail")
)

type OrderCreator interface {
	Create(ctx context.Context, tx *sqlx.Tx, userID int, status model.OrderStatus, number string) error
	FindByNumber(ctx context.Context, tx *sqlx.Tx, number string) (*model.Order, bool)
}

type OrderCreateRequest struct {
	Number string `json:"number" validate:"required,alphanum,gte=3,lte=50"`
}

type OrderCreate struct {
	Rep OrderCreator
}

func NewOrderCreate(rep OrderCreator) *OrderCreate {
	return &OrderCreate{
		Rep: rep,
	}
}

func (ocr *OrderCreate) FromRequest(req *http.Request) error {
	const errMsg = "order create from request:"
	rOrder, err := ocr.validate(req)
	if err != nil {
		return fmt.Errorf("%w %w", ErrOrderCreateValidateFail, err)
	}

	user, err := helper.UserWithContext(req.Context())
	if err != nil {
		return helper.ErrUserNotFound
	}

	err = db.Transaction(func(tx *sqlx.Tx) error {
		order, ok := ocr.Rep.FindByNumber(req.Context(), tx, rOrder.Number)
		if ok {
			if order.UserID == user.ID {
				return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByCurrentUser)
			}

			return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByOtherUser)
		}

		if err := ocr.Rep.Create(req.Context(), tx, user.ID, model.OrderStatusNew, rOrder.Number); err != nil {
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
