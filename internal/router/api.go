package router

import (
	"net/http"

	"github.com/arefev/gophermart/internal/handler"
	"github.com/arefev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func api(mw *middleware.Middleware) http.Handler {
	r := chi.NewRouter()

	userHandler := handler.NewUser(mw.Log, mw.Conf)
	orderHandler := handler.NewOrder(mw.Log)
	balanceHandler := handler.NewBalance(mw.Log)

	r.Route("/user", func(r chi.Router) {
		r.Use(chi_middleware.AllowContentType("application/json"))

		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(mw.Authorized)

			// Сохранение номера заказа
			r.Post("/orders", orderHandler.Save)
			// Получение списка загруженных заказов
			r.Get("/orders", orderHandler.List)

			// Получение текущего баланса
			r.Get("/balance", balanceHandler.Get)
			// Запрос на списание средств
			r.Post("/balance/withdraw", balanceHandler.Withdraw)
			// Получение информации о выводе средств
			r.Get("/withdrawals", balanceHandler.Withdrawals)
		})
	})

	return r
}
