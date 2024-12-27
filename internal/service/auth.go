package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthUserNotFound   = errors.New("user not found")
	ErrAuthJsonDecodeFail = errors.New("json decode fail")
	ErrAuthValidateFail   = errors.New("validate fail")
)

type UserFinder interface {
	FindByLogin(tx *sqlx.Tx, login string) (*model.User, error)
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

func (a *auth) FromRequest(req *http.Request) (string, error) {
	rUser := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&rUser); err != nil {
		return "", fmt.Errorf("auth from request %w: %w", ErrAuthJsonDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(rUser); err != nil {
		return "", fmt.Errorf("auth from request %w: %w", ErrAuthValidateFail, err)
	}

	user, err := a.getUser(rUser.Login)
	if err != nil {
		return "", fmt.Errorf("auth from request get user fail: %w", err)
	}

	if !a.checkPassword(user, rUser.Password) {
		return "", ErrAuthUserNotFound
	}

	token, err := a.GenerateToken(user)
	if err!= nil {
        return "", fmt.Errorf("auth from request generate token fail: %w", err)
    }

	return token, nil
}

func (a *auth) getUser(login string) (*model.User, error) {
	var user *model.User
	var err error

	err = db.Transaction(func(tx *sqlx.Tx) error {
		user, err = a.user.FindByLogin(tx, login)

		if err != nil || user == nil {
			return ErrAuthUserNotFound
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("auth authorize transaction fail: %w", err)
	}

	return user, nil
}

func (a *auth) checkPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (a *auth) GenerateToken(user *model.User) (string, error) {
	duration := time.Minute * time.Duration(a.conf.TokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": user.Login,
		"exp":   time.Now().Add(duration).Unix(),
	})

	strToken, err := token.SignedString([]byte(a.conf.TokenSecret))
	if err != nil {
		return "", fmt.Errorf("generate token fail: %w", err)
	}

	return strToken, nil
}
