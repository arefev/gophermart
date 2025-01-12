package router

import (
	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func New(log *zap.Logger, conf *config.Config) *chi.Mux {
	const compressLevel = 5

	mw := middleware.NewMiddleware(log, conf)
	r := chi.NewRouter()
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Compress(compressLevel, "application/json", "text/html"))

	log.Info("Server started")

	r.Mount("/api", api(&mw))

	return r
}
