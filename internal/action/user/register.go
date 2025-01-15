package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/service"
	"github.com/go-playground/validator/v10"
)

type UserCreateRequest struct {
	Login    string `json:"login" validate:"required,gte=1,lte=20,alphanum"`
	Password string `json:"password" validate:"required,lte=40"`
}

type registerAction struct {
	app *application.App
}

func NewRegisterAction(app *application.App) *registerAction {
	return &registerAction{
		app: app,
	}
}

func (r *registerAction) Handle(req *http.Request) (*UserCreateRequest, error) {
	user := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&user); err != nil {
		return nil, fmt.Errorf("register from request %w: %w", service.ErrRegisterJSONDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(user); err != nil {
		return nil, fmt.Errorf("register from request %w: %w", service.ErrRegisterValidateFail, err)
	}

	s := service.NewUserService(r.app)
	if err := s.Create(req.Context(), user.Login, user.Password); err != nil {
		return nil, fmt.Errorf("register from request save fail: %w", err)
	}

	return &user, nil
}
