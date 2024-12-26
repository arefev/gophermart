package repository

import (
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type User struct {
	log *zap.Logger
	db  *sqlx.DB
}

func NewUser(log *zap.Logger) *User {
	return &User{
		log: log,
		db:  db.Connection(),
	}
}

func (u *User) Exists() bool {
	return false
}

func (u *User) Create(login string, password string) error {
	u.log.Sugar().Infof("login: %s, password: %s", login, password)
	return nil
}
