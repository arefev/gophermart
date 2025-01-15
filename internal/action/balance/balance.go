package balance

import (
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service"
)

type balanceAction struct {
	app *application.App
}

func NewBalanceAction(app *application.App) *balanceAction {
	return &balanceAction{
		app: app,
	}
}

func (b *balanceAction) Handle(r *http.Request) (*model.Balance, error) {
	user, err := service.NewUserService(b.app).Authorized(r.Context())
	if err != nil {
		return nil, service.ErrUserNotAuthorized
	}

	balance, err := service.NewUserBalance(b.app).FindByUserID(r.Context(), user.ID)
	if err != nil {
		return nil, fmt.Errorf("find balance from request fail: %w", err)
	}

	return balance, nil
}
