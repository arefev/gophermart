package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRegisterUserExists = errors.New("user already exists")
	ErrRegisterJsonDecodeFail = errors.New("json decode fail")
	ErrRegisterValidateFail = errors.New("validate fail")
)

type UserCreator interface {
	Exists(tx *sqlx.Tx, login string) bool
	Create(tx *sqlx.Tx, login string, password string) error
}

type UserCreateRequest struct {
	Login    string `json:"login" validate:"required,gte=3,lte=10,alpha"`
	Password string `json:"password" validate:"required,lte=30"`
}

type register struct {
	log  *zap.Logger
	user UserCreator
}

func NewRegister(user UserCreator, log *zap.Logger) *register {
	return &register{
		log:  log,
		user: user,
	}
}

func (r *register) FromRequest(req *http.Request) error {
	user := UserCreateRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&user); err != nil {
		return fmt.Errorf("register from request %w: %w", ErrRegisterJsonDecodeFail, err)
	}

	v := validator.New()
	if err := v.Struct(user); err != nil {
		return fmt.Errorf("register from request %w: %w", ErrRegisterValidateFail, err)
	}

	if err := r.Create(user.Login, user.Password); err != nil {
		return fmt.Errorf("register from request save fail: %w", err)
	}

	return nil
}

func (r *register) Create(login string, password string) error {
	err := db.Transaction(func(tx *sqlx.Tx) error {
		if r.user.Exists(tx, login) {
			return ErrRegisterUserExists
		}

		password, err := r.encryptPassword(password)
		if err != nil {
			return fmt.Errorf("encrypt password fail: %w", err)
		}

		if err := r.user.Create(tx, login, password); err != nil {
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
