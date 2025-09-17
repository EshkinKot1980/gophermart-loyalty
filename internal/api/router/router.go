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
type WithdrawalsService = handler.WithdrawalsService

func New(
	a AuthService,
	o OrderService,
	b BalanceService,
	w WithdrawalsService,
	l Logger,
) *chi.Mux {
	logger := middleware.NewLogger(l)
	authorizer := middleware.NewAuthorizer(a)

	authHandler := handler.NewAuth(a)
	orderHandler := handler.NewOrder(o, l)
	balanceHandler := handler.NewBalance(b, l)
	withdrawalsHandler := handler.NewWithdrawals(w, l)

	router := chi.NewRouter()
	router.Use(logger.Log)
	router.Use(middleware.GzipDecompress)

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
				r.Group(func(r chi.Router) {
					r.Use(middleware.GzipCompress)
					r.Get("/", orderHandler.List)
				})
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler.UserBalance)
				r.Route("/withdraw", func(r chi.Router) {
					r.Post("/", withdrawalsHandler.Withdraw)
				})
			})

			r.Route("/withdrawals", func(r chi.Router) {
				r.Use(middleware.GzipCompress)
				r.Get("/", withdrawalsHandler.List)
			})
		})
	})

	return router
}
