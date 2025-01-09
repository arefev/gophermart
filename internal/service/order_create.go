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
	ErrOrderCreateUploadedByCurrentUser = errors.New("order already uploaded by current user")
	ErrOrderCreateUploadedByOtherUser   = errors.New("order already uploaded by other user")
	ErrOrderCreateValidateFail          = errors.New("validate fail")
)

type OrderCreator interface {
	Create(tx *sqlx.Tx, userID int, status model.OrderStatus, number string) error
	FindByNumber(tx *sqlx.Tx, number string) *model.Order
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
	rOrder := OrderCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&rOrder); err != nil {
		return fmt.Errorf("%s decode fail: %w", errMsg, err)
	}

	v := validator.New()
	if err := v.Struct(rOrder); err != nil {
		return fmt.Errorf("%s %w: %w", errMsg, ErrOrderCreateValidateFail, err)
	}

	user, err := UserWithContext(req.Context())
	if err != nil {
		return fmt.Errorf("%s user not found in context: %w", errMsg, err)
	}

	err = db.Transaction(func(tx *sqlx.Tx) error {
		if order := ocr.Rep.FindByNumber(tx, rOrder.Number); order != nil {
			if order.UserID == user.ID {
				return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByCurrentUser)
			}

			return fmt.Errorf("%s %w", errMsg, ErrOrderCreateUploadedByOtherUser)
		}

		if err := ocr.Rep.Create(tx, user.ID, model.OrderStatusNew, rOrder.Number); err != nil {
			return fmt.Errorf("%s create fail: %w", errMsg, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%s transaction fail: %w", errMsg, err)
	}

	return nil
}
