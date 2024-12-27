package router

import (
	"github.com/arefev/gophermart/internal/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func New(log *zap.Logger, conf *config.Config) *chi.Mux {
	r := chi.NewRouter()
	log.Info("Server started")

	r.Mount("/api", api(log, conf))

	return r
}
