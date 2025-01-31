package model

import "time"

type User struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password" db:"password"`
	ID        int       `json:"id" db:"id"`
}
