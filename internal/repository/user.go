package repository

import "go.uber.org/zap"

type User struct {
	log *zap.Logger
}

func NewUser(log *zap.Logger) *User {
	return &User{
		log: log,
	}
}

func (u *User) Exists() bool { 
	return false
}

func (u *User) Save(login string, password string) error {
	u.log.Sugar().Infof("login: %s, password: %s", login, password)
    return nil
}