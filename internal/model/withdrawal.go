package model

import (
	"time"
)

type Withdrawal struct {
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
	ProcessedAt time.Time `json:"-" db:"processed_at"`
	Sum         float64   `json:"sum" db:"sum"`
	OrderID     int       `json:"-" db:"order_id"`
	ID          int       `json:"-" db:"id"`
}

type WithdrawalWithOrderNumber struct {
	Withdrawal
	Number string `json:"number" db:"number"`
}
