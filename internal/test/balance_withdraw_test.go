package test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/arefev/gophermart/internal/application"
	mock_application "github.com/arefev/gophermart/internal/application/mocks"
	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/model"
	"github.com/arefev/gophermart/internal/router"
	"github.com/arefev/gophermart/internal/service/password"
	"github.com/arefev/gophermart/internal/trm"
	mock_trm "github.com/arefev/gophermart/internal/trm/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestBalanceWithdrawSuccess(t *testing.T) {
	t.Run("balance withdraw success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret:   gofakeit.DigitN(10),
			LogLevel:      "debug",
			TokenDuration: 5,
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, err := password.Encrypt(pwd)
		require.NoError(t, err)

		user := model.User{
			ID:       1,
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		number := "45031620082273"
		withdraw := 100.0
		balance := model.Balance{
			ID:        1,
			UserID:    user.ID,
			Current:   500,
			Withdrawn: 200,
		}

		newCurrent := balance.Current - withdraw
		newWithdrawn := balance.Withdrawn + withdraw

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(2)

		orderRepo := mock_application.NewMockOrderRepo(ctrl)
		orderRepo.EXPECT().CreateWithdrawal(gomock.Any(), user.ID, number, withdraw).MaxTimes(1)

		balanceRepo := mock_application.NewMockBalanceRepo(ctrl)
		balanceRepo.EXPECT().FindByUserID(gomock.Any(), user.ID).Return(&balance, true).MaxTimes(1)
		balanceRepo.EXPECT().UpdateByID(gomock.Any(), balance.ID, newCurrent, newWithdrawn).Return(nil).MaxTimes(1)

		app := application.App{
			Rep: application.Repository{
				User:    userRepo,
				Order:   orderRepo,
				Balance: balanceRepo,
			},
			TrManager: trManager,
			Log:       zLog,
			Conf:      &conf,
		}

		r := router.New(&app)
		srv := httptest.NewServer(r)
		defer srv.Close()

		body := `{
			"login": "` + user.Login + `",
			"password": "` + pwd + `"
		}`

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.Contains(t, hAuth, "Bearer ")

		sum := strconv.FormatFloat(withdraw, 'f', -1, 64)
		body = `{
			"order": "` + number + `",
			"sum": ` + sum + `
		}`

		resp, err = resty.New().
			R().
			SetHeader("Authorization", hAuth).
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/balance/withdraw")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestBalanceWithdrawStatusPayment(t *testing.T) {
	t.Run("balance withdraw status payment", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret:   gofakeit.DigitN(10),
			LogLevel:      "debug",
			TokenDuration: 5,
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, err := password.Encrypt(pwd)
		require.NoError(t, err)

		user := model.User{
			ID:       1,
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		number := "45031620082273"
		withdraw := 100.0
		balance := model.Balance{
			ID:        1,
			UserID:    user.ID,
			Current:   50,
			Withdrawn: 200,
		}

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(2)

		balanceRepo := mock_application.NewMockBalanceRepo(ctrl)
		balanceRepo.EXPECT().FindByUserID(gomock.Any(), user.ID).Return(&balance, true).MaxTimes(1)

		app := application.App{
			Rep: application.Repository{
				User:    userRepo,
				Balance: balanceRepo,
			},
			TrManager: trManager,
			Log:       zLog,
			Conf:      &conf,
		}

		r := router.New(&app)
		srv := httptest.NewServer(r)
		defer srv.Close()

		body := `{
			"login": "` + user.Login + `",
			"password": "` + pwd + `"
		}`

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.Contains(t, hAuth, "Bearer ")

		sum := strconv.FormatFloat(withdraw, 'f', -1, 64)
		body = `{
			"order": "` + number + `",
			"sum": ` + sum + `
		}`

		resp, err = resty.New().
			R().
			SetHeader("Authorization", hAuth).
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/balance/withdraw")

		require.NoError(t, err)
		require.Equal(t, http.StatusPaymentRequired, resp.StatusCode())
	})
}

func TestBalanceWithdrawStatusUnprocessable(t *testing.T) {
	t.Run("balance withdraw status unprocessable", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret:   gofakeit.DigitN(10),
			LogLevel:      "debug",
			TokenDuration: 5,
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, err := password.Encrypt(pwd)
		require.NoError(t, err)

		user := model.User{
			ID:       1,
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		number := "45031620082273"

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(2)

		app := application.App{
			Rep: application.Repository{
				User: userRepo,
			},
			TrManager: trManager,
			Log:       zLog,
			Conf:      &conf,
		}

		r := router.New(&app)
		srv := httptest.NewServer(r)
		defer srv.Close()

		body := `{
			"login": "` + user.Login + `",
			"password": "` + pwd + `"
		}`

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.Contains(t, hAuth, "Bearer ")

		body = `{
			"order": "` + number + `"
		}`

		resp, err = resty.New().
			R().
			SetHeader("Authorization", hAuth).
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/balance/withdraw")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode())
	})
}
