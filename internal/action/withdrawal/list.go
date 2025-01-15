package withdrawal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service"
)

type listAction struct {
	app  *application.App
	list []model.Order
}

func NewListAction(app *application.App) *listAction {
	return &listAction{
		app:  app,
		list: make([]model.Order, 0),
	}
}

func (l *listAction) Handle(r *http.Request) ([]model.Withdrawal, error) {
	user, err := service.NewUserService(l.app).Authorized(r.Context())
	if err != nil {
		return []model.Withdrawal{}, service.ErrUserNotAuthorized
	}

	var list []model.Withdrawal
	err = l.app.TrManager.Do(r.Context(), func(ctx context.Context) error {
		list = l.app.Rep.Order.GetWithdrawalsByUserID(ctx, user.ID)

		return nil
	})

	if err != nil {
		return []model.Withdrawal{}, fmt.Errorf("withdrawal list transaction fail: %w", err)
	}

	return list, nil
}
