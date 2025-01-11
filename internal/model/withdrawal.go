package model

import (
	"time"
)

type Withdrawal struct {
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
	ProcessedAt time.Time `json:"-" db:"processed_at"`
	Number      string    `json:"number" db:"number"`
	Sum         float64   `json:"sum" db:"sum"`
	UserID      int       `json:"-" db:"user_id"`
	ID          int       `json:"-" db:"id"`
}
