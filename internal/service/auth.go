package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service/jwt"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthUserNotFound   = errors.New("user not found")
	ErrAuthJSONDecodeFail = errors.New("json decode fail")
	ErrAuthValidateFail   = errors.New("validate fail")
)

type UserAuthRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type auth struct {
	app *application.App
}

func NewAuth(app *application.App) *auth {
	return &auth{
		app: app,
	}
}

func (a *auth) FromRequest(req *http.Request) (*jwt.Token, error) {
	rUser := UserAuthRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", ErrAuthJSONDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", ErrAuthValidateFail, err)
	}

	token, err := a.Authorize(req.Context(), rUser.Login, rUser.Password)
	if err != nil {
		return nil, fmt.Errorf("auth from request fail: %w", err)
	}

	return token, nil
}

func (a *auth) Authorize(ctx context.Context, login, password string) (*jwt.Token, error) {
	user, err := a.GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("authorize get user fail: %w", err)
	}

	if !a.checkPassword(user, password) {
		return nil, ErrAuthUserNotFound
	}

	token, err := jwt.NewToken(a.app.Conf.TokenSecret).GenerateToken(user, a.app.Conf.TokenDuration)
	if err != nil {
		return nil, fmt.Errorf("auth from request generate token fail: %w", err)
	}

	return token, nil
}

func (a *auth) GetUser(ctx context.Context, login string) (*model.User, error) {
	var user *model.User
	var ok bool

	err := a.app.TrManager.Do(ctx, func(ctx context.Context) error {
		user, ok = a.app.Rep.User.FindByLogin(ctx, login)

		if !ok {
			return ErrAuthUserNotFound
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("get user transaction fail: %w", err)
	}

	return user, nil
}

func (a *auth) checkPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
