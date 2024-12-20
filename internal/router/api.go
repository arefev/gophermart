package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func API(log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	log.Info("Server started")

	return r
}
