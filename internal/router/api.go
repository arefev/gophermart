package router

import (
	"net/http"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/handler"
	"github.com/arefev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func api(app *application.App, mw *middleware.Middleware) http.Handler {
	r := chi.NewRouter()
	r.Use(chi_middleware.AllowContentType("application/json", "text/plain"))
	r.Use(chi_middleware.SetHeader("Content-Type", "application/json"))

	userHandler := handler.NewUser(app)
	orderHandler := handler.NewOrder(app)
	balanceHandler := handler.NewBalance(app)

	r.Route("/user", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(mw.Authorized)

			// Сохранение номера заказа
			r.Post("/orders", orderHandler.Create)
			// Получение списка загруженных заказов
			r.Get("/orders", orderHandler.List)

			// Получение текущего баланса
			r.Get("/balance", balanceHandler.Find)
			// Запрос на списание средств
			r.Post("/balance/withdraw", balanceHandler.Withdraw)
			// Получение информации о выводе средств
			r.Get("/withdrawals", balanceHandler.Withdrawals)
		})
	})

	return r
}
