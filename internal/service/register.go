package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserCreator interface {
	Exists() bool
	Create(login string, password string) error
}

type UserCreateRequest struct {
	Login    string `json:"login" validate:"required,lte=10,alpha"`
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
		return fmt.Errorf("register from request json decode fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(user); err != nil {
		return fmt.Errorf("register from request validate fail: %w", err)
	}

	if err := r.Create(user.Login, user.Password); err != nil {
		return fmt.Errorf("register from request save fail: %w", err)
	}

	return nil
}

func (r *register) Create(login string, password string) error {
	if r.user.Exists() {
		return fmt.Errorf("register create fail: user already exists")
	}

	password, err := r.encryptPassword(password)
	if err != nil {
		return fmt.Errorf("register create fail: %w", err)
	}

	if err := r.user.Create(login, password); err != nil {
		return fmt.Errorf("register create fail: %w", err)
	}

	return nil
}

func (r *register) encryptPassword(password string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("encrypt password fail: %w", err)
	}

	return string(passHash), nil
}
