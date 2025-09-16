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
	Register(ctx context.Context, c dto.Credentials) (token string, err error)
	Login(ctx context.Context, c dto.Credentials) (token string, err error)
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
		http.Error(w, "invalid credentials format: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), credentials)
	if err != nil {
		switch {
		case errors.Is(err, srvErrors.ErrAuthInvalidCredentials):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, srvErrors.ErrAuthUserAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var credentials dto.Credentials

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "invalid credentials format: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), credentials)
	if err != nil {
		if errors.Is(err, srvErrors.ErrAuthInvalidCredentials) {
			http.Error(w, "", http.StatusUnauthorized)
		} else {
			http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
