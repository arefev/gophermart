package router

import (
	"net/http"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/handler"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func api(log *zap.Logger, conf *config.Config) http.Handler {
	r := chi.NewRouter()

	userHandler := handler.NewUser(log, conf)
	orderHandler := handler.NewOrder(log)
	balanceHandler := handler.NewBalance(log)

	r.Route("/user", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)

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

	return r
}
