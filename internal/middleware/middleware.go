package middleware

import (
	"github.com/arefev/gophermart/internal/config"
	"go.uber.org/zap"
)

type Middleware struct {
	Log  *zap.Logger
	Conf *config.Config
}

func NewMiddleware(log *zap.Logger, conf *config.Config) Middleware {
	return Middleware{
		Log:  log,
		Conf: conf,
	}
}
