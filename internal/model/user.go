package model

import "time"

type User struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password" db:"password"`
	ID        int       `json:"id" db:"id"`
}
