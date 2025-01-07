package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/arefev/gophermart/internal/service/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthUserNotFound   = errors.New("user not found")
	ErrAuthJSONDecodeFail = errors.New("json decode fail")
	ErrAuthValidateFail   = errors.New("validate fail")
)

type UserFinder interface {
	FindByLogin(tx *sqlx.Tx, login string) *model.User
}

type UserAuthRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type auth struct {
	log  *zap.Logger
	user UserFinder
	conf *config.Config
}

func NewAuth(user UserFinder, log *zap.Logger, conf *config.Config) *auth {
	return &auth{
		user: user,
		log:  log,
		conf: conf,
	}
}

func (a *auth) FromRequest(req *http.Request) (*jwt.Token, error) {
	rUser := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", ErrAuthJSONDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(rUser); err != nil {
		return nil, fmt.Errorf("auth from request %w: %w", ErrAuthValidateFail, err)
	}

	token, err := a.Authorize(rUser.Login, rUser.Password)
	if err != nil {
		return nil, fmt.Errorf("auth from request fail: %w", err)
	}

	return token, nil
}

func (a *auth) Authorize(login, password string) (*jwt.Token, error) {
	user, err := a.GetUser(login)
	if err != nil {
		return nil, fmt.Errorf("authorize get user fail: %w", err)
	}

	if !a.checkPassword(user, password) {
		return nil, ErrAuthUserNotFound
	}

	token, err := jwt.NewToken(a.conf.TokenSecret).GenerateToken(user, a.conf.TokenDuration)
	if err != nil {
		return nil, fmt.Errorf("auth from request generate token fail: %w", err)
	}

	return token, nil
}

func (a *auth) GetUser(login string) (*model.User, error) {
	var user *model.User

	err := db.Transaction(func(tx *sqlx.Tx) error {
		user = a.user.FindByLogin(tx, login)

		if user == nil {
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
