package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/service"
	"github.com/arefev/gophermart/internal/service/jwt"
	"github.com/go-playground/validator/v10"
)

type UserAuthRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type authAction struct {
	app *application.App
}

func NewAuthAction(app *application.App) *authAction {
	return &authAction{
		app: app,
	}
}

func (a *authAction) Handle(r *http.Request) (*jwt.Token, error) {
	rUser := UserAuthRequest{}
	d := json.NewDecoder(r.Body)

	if err := d.Decode(&rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", service.ErrAuthJSONDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", service.ErrAuthValidateFail, err)
	}

	s := service.NewUserService(a.app)
	token, err := s.Authorize(r.Context(), rUser.Login, rUser.Password)
	if err != nil {
		return nil, fmt.Errorf("auth from request fail: %w", err)
	}

	return token, nil
}
