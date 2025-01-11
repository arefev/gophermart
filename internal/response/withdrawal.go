package response

import (
	"time"

	"github.com/arefev/gophermart/internal/model"
)

type Withdrawal struct {
	ProcessedAt time.Time `json:"processed_at"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
}

func NewWithdrawal(w *model.WithdrawalWithOrderNumber) Withdrawal {
	return Withdrawal{
		Order:       w.Number,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}
}

func NewWithdrawals(l []model.WithdrawalWithOrderNumber) *[]Withdrawal {
	withdrawals := make([]Withdrawal, 0, len(l))
	for i := range l {
		withdrawals = append(withdrawals, NewWithdrawal(&l[i]))
	}
	return &withdrawals
}
