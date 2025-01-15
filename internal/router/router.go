package router

import (
	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func New(app *application.App) *chi.Mux {
	const compressLevel = 5

	mw := middleware.NewMiddleware(app)
	r := chi.NewRouter()
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Compress(compressLevel, "application/json", "text/html"))

	app.Log.Info("Server started")

	r.Mount("/api", api(app, &mw))

	return r
}
