package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler"
)

func New() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", handler.InfoPage)

	return router
}
