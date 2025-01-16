package test

import (
	"net/http"
	"net/http/httptest"
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

func TestOrderCreateSuccess(t *testing.T) {
	t.Run("order create success", func(t *testing.T) {
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

		orderNumber := "45031620082273"

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(2)

		orderRepo := mock_application.NewMockOrderRepo(ctrl)
		orderRepo.EXPECT().FindByNumber(gomock.Any(), orderNumber).Return(nil, false).MaxTimes(1)
		orderRepo.EXPECT().Create(gomock.Any(), user.ID, model.OrderStatusNew, orderNumber).Return(nil).MaxTimes(1)

		app := application.App{
			Rep: application.Repository{
				User:  userRepo,
				Order: orderRepo,
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

		resp, err = resty.New().
			R().
			SetHeader("Content-type", "text/plain").
			SetHeader("Authorization", hAuth).
			SetBody(orderNumber).
			Post(srv.URL + "/api/user/orders")

		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode())
	})
}

func TestOrderCreateStatusUnprocessable(t *testing.T) {
	t.Run("order create status unprocessable", func(t *testing.T) {
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

		orderNumber := "1234"

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

		resp, err = resty.New().
			R().
			SetHeader("Content-type", "text/plain").
			SetHeader("Authorization", hAuth).
			SetBody(orderNumber).
			Post(srv.URL + "/api/user/orders")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode())
	})
}

func TestOrderCreateStatusUnAuthorized(t *testing.T) {
	t.Run("order create status unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret:   gofakeit.DigitN(10),
			LogLevel:      "debug",
			TokenDuration: 5,
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		orderNumber := "45031620082273"

		app := application.App{
			Log:  zLog,
			Conf: &conf,
		}

		r := router.New(&app)
		srv := httptest.NewServer(r)
		defer srv.Close()

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "text/plain").
			SetBody(orderNumber).
			Post(srv.URL + "/api/user/orders")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
	})
}
