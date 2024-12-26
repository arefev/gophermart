package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
}

func NewAuth(user UserFinder, log *zap.Logger) *auth {
	return &auth{
		user: user,
		log:  log,
	}
}

func (a *auth) FromRequest(req *http.Request) error {
	rUser := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&rUser); err != nil {
		return fmt.Errorf("auth from request json decode fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(rUser); err != nil {
		return fmt.Errorf("auth from request validate fail: %w", err)
	}

	user, err := a.getUser(rUser.Login)
	if err != nil {
		return fmt.Errorf("auth from request get user fail: %w", err)
	}

	if !a.checkPassword(user, rUser.Password) {
		return fmt.Errorf("auth from request wrong pair login / password")
	}

	return nil
}

func (a *auth) getUser(login string) (*model.User, error) {
	var user *model.User
	var err error

	err = db.Transaction(func(tx *sqlx.Tx) error {
		user, err = a.user.FindByLogin(tx, login)

		if err != nil || user == nil {
			return fmt.Errorf("user not found")
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
