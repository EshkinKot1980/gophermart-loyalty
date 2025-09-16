package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/auth"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/order"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
)

type Logger = middleware.HTTPloger
type AuthService interface {
	auth.AuthService
	middleware.AuthService
}
type OrderService = order.OrderService

func New(a AuthService, o OrderService, l Logger) *chi.Mux {
	logger := middleware.NewLogger(l)
	authorizer := middleware.NewAuthorizer(a)

	authHandler := auth.New(a)
	orderHandler := order.New(o)

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
			})
		})
	})

	return router
}
