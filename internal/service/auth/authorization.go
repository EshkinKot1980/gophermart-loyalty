package auth

import (
	"context"
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

func (a *Auth) User(ctx context.Context, token string) (entity.User, error) {
	var user entity.User

	jt, err := jwt.Parse(
		token,
		func(t *jwt.Token) (any, error) {
			return []byte(a.secret), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return user, srvErrors.ErrAuthTokenExpired
		}
		return user, srvErrors.ErrAuthInvalidToken
	}

	userID, err := parseID(jt)
	if err != nil {
		return user, srvErrors.ErrAuthInvalidToken
	}

	user, err = a.repository.GetByID(ctx, userID)
	if err != nil {
		return user, srvErrors.ErrAuthInvalidToken
	}

	return user, nil
}

func parseID(token *jwt.Token) (uint64, error) {
	var id uint64

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return id, srvErrors.ErrAuthInvalidToken
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return id, srvErrors.ErrAuthInvalidToken
	}

	id, err := strconv.ParseUint(jti, 10, 64)
	if err != nil {
		return id, srvErrors.ErrAuthInvalidToken
	}

	return id, nil
}
