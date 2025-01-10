package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRegisterUserExists     = errors.New("user already exists")
	ErrRegisterJSONDecodeFail = errors.New("json decode fail")
	ErrRegisterValidateFail   = errors.New("validate fail")
)

type UserCreator interface {
	Exists(tx *sqlx.Tx, login string) bool
	Create(tx *sqlx.Tx, login string, password string) error
}

type UserCreateRequest struct {
	Login    string `json:"login" validate:"required,gte=1,lte=20,alphanum"`
	Password string `json:"password" validate:"required,lte=40"`
}

type register struct {
	user UserCreator
	conf *config.Config
}

func NewRegister(user UserCreator, conf *config.Config) *register {
	return &register{
		user: user,
		conf: conf,
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

	if err := r.Create(user.Login, user.Password); err != nil {
		return nil, fmt.Errorf("register from request save fail: %w", err)
	}

	return &user, nil
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
