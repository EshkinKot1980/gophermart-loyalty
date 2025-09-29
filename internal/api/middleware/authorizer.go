package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type AuthService interface {
	User(ctx context.Context, token string) (entity.User, error)
}

type Authorizer struct {
	service AuthService
}

type ContextKey string

const KeyUserID ContextKey = "userID"

func NewAuthorizer(srv AuthService) *Authorizer {
	return &Authorizer{service: srv}
}

func (a *Authorizer) Authorize(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		user, err := a.service.User(r.Context(), token)
		if err != nil {
			if errors.Is(err, srvErrors.ErrAuthTokenExpired) {
				http.Error(w, "token expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "", http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), KeyUserID, user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
