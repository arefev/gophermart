package middleware

import (
	"github.com/arefev/gophermart/internal/application"
)

type Middleware struct {
	app *application.App
}

func NewMiddleware(app *application.App) Middleware {
	return Middleware{
		app: app,
	}
}
