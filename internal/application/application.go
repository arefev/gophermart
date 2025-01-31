package application

import (
	"context"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/trm"
	"go.uber.org/zap"
)

type UserRepo interface {
	Exists(ctx context.Context, login string) bool
	FindByLogin(ctx context.Context, login string) (*model.User, bool)
	Create(ctx context.Context, login, password string) error
}

type OrderRepo interface {
	FindByNumber(ctx context.Context, number string) (*model.Order, bool)
	Create(ctx context.Context, userID int, status model.OrderStatus, number string) error
	GetByUserID(ctx context.Context, userID int) []model.Order
	WithStatusNew(ctx context.Context) []model.Order
	AccrualByID(ctx context.Context, sum float64, status model.OrderStatus, id int) error
	CreateWithdrawal(ctx context.Context, userID int, number string, sum float64) error
	GetWithdrawalsByUserID(ctx context.Context, userID int) []model.Withdrawal
}

type BalanceRepo interface {
	FindByUserID(ctx context.Context, userID int) (*model.Balance, bool)
	UpdateByID(ctx context.Context, id int, current, withdrawn float64) error
}

type TrManager interface {
	Do(ctx context.Context, action trm.TrAction) error
}

type App struct {
	Rep       Repository
	TrManager TrManager
	Log       *zap.Logger
	Conf      *config.Config
}

type Repository struct {
	User    UserRepo
	Order   OrderRepo
	Balance BalanceRepo
}
