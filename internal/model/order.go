package model

import "time"

type Order struct {
	CreatedAt  time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time   `json:"updatedAt" db:"updated_at"`
	UploadedAt time.Time   `json:"uploadedAt" db:"uploaded_at"`
	Number     string      `json:"number" db:"number"`
	Status     OrderStatus `json:"status" db:"status"`
	UserID     int         `json:"userId" db:"user_id"`
	ID         int         `json:"id" db:"id"`
}

type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota + 1
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusProcessing:
		return "PROCESSING"
	case OrderStatusInvalid:
		return "INVALID"
	case OrderStatusProcessed:
		return "PROCESSED"
	default:
		return "NEW"
	}
}
