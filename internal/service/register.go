package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRegisterUserExists     = errors.New("user already exists")
	ErrRegisterJSONDecodeFail = errors.New("json decode fail")
	ErrRegisterValidateFail   = errors.New("validate fail")
)

type UserCreateRequest struct {
	Login    string `json:"login" validate:"required,gte=1,lte=20,alphanum"`
	Password string `json:"password" validate:"required,lte=40"`
}

type register struct {
	app *application.App
}

func NewRegister(app *application.App) *register {
	return &register{
		app: app,
	}
}

func (r *register) FromRequest(req *http.Request) (*UserCreateRequest, error) {
	user := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&user); err != nil {
		return nil, fmt.Errorf("register from request %w: %w", ErrRegisterJSONDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(user); err != nil {
		return nil, fmt.Errorf("register from request %w: %w", ErrRegisterValidateFail, err)
	}

	if err := r.Create(req.Context(), user.Login, user.Password); err != nil {
		return nil, fmt.Errorf("register from request save fail: %w", err)
	}

	return &user, nil
}

func (r *register) Create(ctx context.Context, login string, password string) error {
	err := r.app.TrManager.Do(ctx, func(ctx context.Context) error {
		if r.app.Rep.User.Exists(ctx, login) {
			return ErrRegisterUserExists
		}

		password, err := r.encryptPassword(password)
		if err != nil {
			return fmt.Errorf("encrypt password fail: %w", err)
		}

		if err := r.app.Rep.User.Create(ctx, login, password); err != nil {
			return fmt.Errorf("create user fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("register create transaction fail: %w", err)
	}

	return nil
}

func (r *register) encryptPassword(password string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("register encrypt password fail: %w", err)
	}

	return string(passHash), nil
}
