package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/auth"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
)

type Logger = middleware.HTTPloger
type AuthService = auth.AuthService

func New(a AuthService, l Logger) *chi.Mux {
	mwLogger := middleware.NewLogWraper(l)

	authHandler := auth.New(a)

	router := chi.NewRouter()
	router.Use(mwLogger.Log)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
	})

	return router
}
