package model

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
