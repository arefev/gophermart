package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type UserSaver interface {
	Exists() bool
	Save(login string, password string) error
}

type UserRequest struct {
	Login    string `json:"login" validate:"required,lte=10,alpha"`
	Password string `json:"password" validate:"required,lte=30"`
}

type register struct {
	log  *zap.Logger
	user UserSaver
}

func NewRegister(user UserSaver, log *zap.Logger) *register {
	return &register{
		log:  log,
		user: user,
	}
}

func (r *register) FromRequest(req *http.Request) error {
	user := UserRequest{}
	d := json.NewDecoder(req.Body)

	if err := d.Decode(&user); err != nil {
		return fmt.Errorf("register from request json decode fail: %w", err)
	}

	v := validator.New()
	if err := v.Struct(user); err != nil {
		return fmt.Errorf("register from request validate fail: %w", err)
	}

	if err := r.Save(user.Login, user.Password); err != nil {
		return fmt.Errorf("register from request save fail: %w", err)
	}

	return nil
}

func (r *register) Save(login string, password string) error {
	if r.user.Exists() {
		return fmt.Errorf("register save fail: user already exists")
	}

	password, err := r.encryptPassword(password)
	if err != nil {
		return fmt.Errorf("register save fail: encrypt password fail: %w", err)
	}

	if err := r.user.Save(login, password); err != nil {
		return fmt.Errorf("register save fail: %w", err)
	}

	return nil
}

func (r *register) encryptPassword(password string) (string, error) {
	return password, nil
}
