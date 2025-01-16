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

func TestUserAuthSuccess(t *testing.T) {
	t.Run("authorize success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret: gofakeit.DigitN(10),
			LogLevel:    "debug",
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, _ := password.Encrypt(pwd)

		user := model.User{
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(1)

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
	})
}

func TestUserAuthStatusUnauth(t *testing.T) {
	t.Run("authorize status unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret: gofakeit.DigitN(10),
			LogLevel:    "debug",
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		otherPwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, _ := password.Encrypt(pwd)

		user := model.User{
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(1)

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
			"password": "` + otherPwd + `"
		}`

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.NotContains(t, hAuth, "Bearer ")
	})
}

func TestUserAuthStatusBadRequest(t *testing.T) {
	t.Run("authorize status bad request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret: gofakeit.DigitN(10),
			LogLevel:    "debug",
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, _ := password.Encrypt(pwd)

		user := model.User{
			Login:    gofakeit.Username(),
			Password: pwdHash,
		}

		tr := mock_trm.NewMockTransaction(ctrl)
		trManager := trm.NewTrm(tr, zLog)
		tr.EXPECT().Begin(gomock.Any()).AnyTimes()
		tr.EXPECT().Commit(gomock.Any()).AnyTimes()
		tr.EXPECT().Rollback(gomock.Any()).AnyTimes()

		userRepo := mock_application.NewMockUserRepo(ctrl)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(1)

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
			"login": "` + user.Login + `"
		}`

		resp, err := resty.New().
			R().
			SetHeader("Content-type", "application/json").
			SetBody(body).
			Post(srv.URL + "/api/user/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.NotContains(t, hAuth, "Bearer ")
	})
}