package router

import (
	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func New(log *zap.Logger, conf *config.Config) *chi.Mux {
	mw := middleware.NewMiddleware(log, conf)
	r := chi.NewRouter()
	log.Info("Server started")

	r.Mount("/api", api(&mw))

	return r
}
