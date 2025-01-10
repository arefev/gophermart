package model

import "time"

type Balance struct {
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
	Current   float64   `json:"current" db:"current"`
	Withdrawn float64   `json:"withdrawn" db:"withdrawn"`
	UserID    int       `json:"-" db:"user_id"`
	ID        int       `json:"-" db:"id"`
}
