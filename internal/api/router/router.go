package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
)

type Logger interface {
	middleware.HTTPloger
	handler.Logger
}
type AuthService interface {
	handler.AuthService
	middleware.AuthService
}
type OrderService = handler.OrderService
type BalanceService = handler.BalanceService

func New(a AuthService, o OrderService, b BalanceService, l Logger) *chi.Mux {
	logger := middleware.NewLogger(l)
	authorizer := middleware.NewAuthorizer(a)

	authHandler := handler.NewAuth(a)
	orderHandler := handler.NewOrder(o, l)
	balanceHandler := handler.NewBalance(b, l)

	router := chi.NewRouter()
	router.Use(logger.Log)

	router.Route("/api/user", func(r chi.Router) {
		r.Route("/register", func(r chi.Router) {
			r.Post("/", authHandler.Register)
		})

		r.Route("/login", func(r chi.Router) {
			r.Post("/", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(authorizer.Authorize)

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.Create)
				r.Get("/", orderHandler.List)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler.UserBalance)
			})
		})
	})

	return router
}
