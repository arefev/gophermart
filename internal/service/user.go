package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/service/jwt"
	"github.com/arefev/gophermart/internal/service/password"
)

var (
	ErrRegisterUserExists     = errors.New("user already exists")
	ErrRegisterJSONDecodeFail = errors.New("json decode fail")
	ErrRegisterValidateFail   = errors.New("validate fail")
	ErrAuthUserNotFound       = errors.New("user not found")
	ErrAuthJSONDecodeFail     = errors.New("json decode fail")
	ErrAuthValidateFail       = errors.New("validate fail")
	ErrUserNotAuthorized      = errors.New("user not authorized")
)

type userService struct {
	app *application.App
}

func NewUserService(app *application.App) *userService {
	return &userService{
		app: app,
	}
}

func (us *userService) Create(ctx context.Context, login string, pwd string) error {
	err := us.app.TrManager.Do(ctx, func(ctx context.Context) error {
		if us.app.Rep.User.Exists(ctx, login) {
			return ErrRegisterUserExists
		}

		pwdHash, err := password.Encrypt(pwd)
		if err != nil {
			return fmt.Errorf("encrypt password fail: %w", err)
		}

		if err := us.app.Rep.User.Create(ctx, login, pwdHash); err != nil {
			return fmt.Errorf("create user fail: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("register create transaction fail: %w", err)
	}

	return nil
}

func (us *userService) Authorize(ctx context.Context, login, pwd string) (*jwt.Token, error) {
	user, err := us.GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("authorize get user fail: %w", err)
	}

	if !password.Check(user.Password, pwd) {
		return nil, ErrAuthUserNotFound
	}

	token, err := jwt.NewToken(us.app.Conf.TokenSecret).GenerateToken(user, us.app.Conf.TokenDuration)
	if err != nil {
		return nil, fmt.Errorf("auth from request generate token fail: %w", err)
	}

	return token, nil
}

func (us *userService) GetUser(ctx context.Context, login string) (*model.User, error) {
	var user *model.User
	var ok bool

	err := us.app.TrManager.Do(ctx, func(ctx context.Context) error {
		user, ok = us.app.Rep.User.FindByLogin(ctx, login)
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

func (us *userService) Authorized(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(model.User{}).(*model.User)

	if !ok {
		return nil, errors.New("user not authorized")
	}

	return user, nil
}
