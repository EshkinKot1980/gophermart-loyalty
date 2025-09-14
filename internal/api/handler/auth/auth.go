package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type AuthService interface {
	Register(ctx context.Context, c dto.Credentials) error
	// Login(ctx context.Context, c dto.Credentials) error
}

type Auth struct {
	service AuthService
}

func New(srv AuthService) *Auth {
	return &Auth{service: srv}
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var credentials dto.Credentials

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, ""+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Register(r.Context(), credentials); err != nil {
		switch {
		case errors.Is(err, srvErrors.ErrInvalidCredentials):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, srvErrors.ErrUserAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
