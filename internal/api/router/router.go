package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/auth"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
)

type Logger = middleware.HTTPloger
type AuthService interface {
	auth.AuthService
	middleware.AuthService
}

func New(a AuthService, l Logger) *chi.Mux {
	mwLogger := middleware.NewLogger(l)
	// mwAuth := middleware.NewAuthorizer(a)

	authHandler := auth.New(a)

	router := chi.NewRouter()
	router.Use(mwLogger.Log)

	router.Route("/api/user", func(r chi.Router) {
		r.Route("/register", func(r chi.Router) {
			r.Post("/", authHandler.Register)
		})
		r.Route("/login", func(r chi.Router) {
			r.Post("/", authHandler.Login)
		})
	})

	return router
}
