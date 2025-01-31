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

func TestUserRegisterSuccess(t *testing.T) {
	type want struct {
		userExists bool
		status     int
		hValue     string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "register success",
			want: want{
				userExists: false,
				status:     http.StatusOK,
				hValue:     "Bearer ",
			},
		},
		{
			name: "register status conflict",
			want: want{
				userExists: true,
				status:     http.StatusConflict,
				hValue:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conf := config.Config{
				TokenSecret: gofakeit.DigitN(10),
				LogLevel:    "debug",
			}

			zLog, err := logger.Build(conf.LogLevel)
			require.NoError(t, err)

			pwd := gofakeit.Password(true, true, true, true, false, 10)
			pwdHash, err := password.Encrypt(pwd)
			require.NoError(t, err)

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
			userRepo.EXPECT().Exists(gomock.Any(), user.Login).Return(tt.want.userExists).MaxTimes(1)
			userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(1)
			userRepo.EXPECT().Create(gomock.Any(), user.Login, gomock.Any()).Return(nil).MaxTimes(1)

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
				Post(srv.URL + "/api/user/register")

			require.NoError(t, err)
			require.Equal(t, tt.want.status, resp.StatusCode())

			hAuth := resp.Header().Get("Authorization")
			require.Contains(t, hAuth, tt.want.hValue)
		})
	}
}

func TestUserRegisterStatusBadRequest(t *testing.T) {
	t.Run("register status bad request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conf := config.Config{
			TokenSecret: gofakeit.DigitN(10),
			LogLevel:    "debug",
		}

		zLog, err := logger.Build(conf.LogLevel)
		require.NoError(t, err)

		pwd := gofakeit.Password(true, true, true, true, false, 10)
		pwdHash, err := password.Encrypt(pwd)
		require.NoError(t, err)

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
		userRepo.EXPECT().Exists(gomock.Any(), user.Login).Return(false).MaxTimes(1)
		userRepo.EXPECT().FindByLogin(gomock.Any(), user.Login).Return(&user, true).MaxTimes(1)
		userRepo.EXPECT().Create(gomock.Any(), user.Login, gomock.Any()).Return(nil).MaxTimes(1)

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
			Post(srv.URL + "/api/user/register")

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())

		hAuth := resp.Header().Get("Authorization")
		require.NotContains(t, hAuth, "Bearer ")
	})
}
