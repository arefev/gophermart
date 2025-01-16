package test

import (
	"context"
	"testing"

	"github.com/arefev/gophermart/internal/application"
	mock_application "github.com/arefev/gophermart/internal/application/mocks"
	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/trm"
	mock_trm "github.com/arefev/gophermart/internal/trm/mocks"
	"github.com/arefev/gophermart/internal/worker"
	mock_worker "github.com/arefev/gophermart/internal/worker/mocks"
	"github.com/golang/mock/gomock"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestWorkerSuccess(t *testing.T) {
	t.Run("authorize success", func(t *testing.T) {
		ctx := context.Background()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret:  gofakeit.DigitN(10),
			PollInterval: 2,
			LogLevel:     "debug",
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		user := model.User{
			ID: 1,
		}

		number := "45031620082273"
		accrual := 100.0
		balance := model.Balance{
			ID:        1,
			UserID:    user.ID,
			Current:   500,
			Withdrawn: 200,
		}

		res := worker.OrderResponse{}
		order := model.Order{
			ID:     1,
			UserID: 1,
			Number: number,
			Status: model.OrderStatusNew,
		}

		newOrders := []model.Order{order}

		newCurrent := balance.Current + accrual
		newStatus := model.OrderStatusProcessed

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		balanceRepo := mock_application.NewMockBalanceRepo(ctrl)
		balanceRepo.EXPECT().FindByUserID(gomock.Any(), user.ID).Return(&balance, true).MaxTimes(1)
		balanceRepo.EXPECT().UpdateByID(gomock.Any(), balance.ID, newCurrent, balance.Withdrawn).Return(nil).MaxTimes(1)

		orderRepo := mock_application.NewMockOrderRepo(ctrl)
		orderRepo.EXPECT().WithStatusNew(gomock.Any()).Return(newOrders).MaxTimes(1)
		orderRepo.EXPECT().AccrualByID(gomock.Any(), accrual, newStatus, order.ID).Return(nil).MaxTimes(1)

		r := mock_worker.NewMockStatusRequest(ctrl)
		r.EXPECT().Request(gomock.Any(), order.Number, &res).Do(func(ctx context.Context, number string, res *worker.OrderResponse) {
			res.Status = newStatus.String()
			res.Accrual = accrual
		})

		app := application.App{
			Rep: application.Repository{
				Order:   orderRepo,
				Balance: balanceRepo,
			},
			TrManager: trManager,
			Log:       zLog,
			Conf:      &conf,
		}

		worker.NewWorker(&app, r).Handle(ctx)
	})
}
