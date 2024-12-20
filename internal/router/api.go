package router

import (
	"github.com/arefev/gophermart/internal/handler"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func API(log *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	log.Info("Server started")

	userHandler := handler.NewUser(log)
	orderHandler := handler.NewOrder(log)
	balanceHandler := handler.NewBalance(log)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)

			r.Post("/orders", orderHandler.Save)
			r.Get("/orders", orderHandler.List)

			r.Get("/balance", balanceHandler.Get)
			r.Post("/balance/withdraw", balanceHandler.Withdraw)
			r.Get("/withdrawals", balanceHandler.Withdrawals)
		})
	})

	return r
}
