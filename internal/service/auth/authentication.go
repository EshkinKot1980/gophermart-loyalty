package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

func (a *Auth) Register(ctx context.Context, c dto.Credentials) (token string, err error) {
	cr := trimCredentials(c)
	if err := validateCredentials(cr); err != nil {
		return token, err
	}

	// bcrypt имеет недостатки, в дальнейшем планирую переделать на другой алгоритм
	hash, err := bcrypt.GenerateFromPassword([]byte(cr.Password), bcrypt.DefaultCost)
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrPasswordTooLong):
			err = fmt.Errorf("%w: password too long, max 72 bytes", srvErrors.ErrInvalidCredentials)
			return token, err
		default:
			a.logger.Error("failed to hash password", err)
			return token, srvErrors.ErrUnexpected
		}
	}

	user := entity.User{Login: c.Login, Hash: string(hash)}
	user, err = a.repository.Create(ctx, user)

	if err != nil {
		switch {
		case errors.Is(err, repErrors.ErrDuplicateKey):
			return token, srvErrors.ErrUserAlreadyExists
		default:
			a.logger.Error("failed to create user", err)
			return token, srvErrors.ErrUnexpected
		}
	}

	return a.generateToken(user)
}

func (a *Auth) Login(ctx context.Context, c dto.Credentials) (token string, err error) {
	cr := trimCredentials(c)

	user, err := a.repository.FindByLogin(ctx, cr.Login)
	if err != nil {
		if errors.Is(err, repErrors.ErrNotFound) {
			return token, srvErrors.ErrInvalidCredentials
		} else {
			a.logger.Error("failed to find user", err)
			return token, srvErrors.ErrUnexpected
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(cr.Password))
	if err != nil {
		return token, srvErrors.ErrInvalidCredentials
	}

	return a.generateToken(user)
}

func (a *Auth) generateToken(u entity.User) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":  u.ID,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		},
	)

	tokenStr, err := token.SignedString([]byte(a.secret))
	if err != nil {
		a.logger.Error("failed to generate token", err)
		return "", srvErrors.ErrUnexpected
	}

	return tokenStr, nil
}

func trimCredentials(c dto.Credentials) dto.Credentials {
	return dto.Credentials{
		Login:    strings.TrimSpace(c.Login),
		Password: strings.TrimSpace(c.Password),
	}
}

func validateCredentials(c dto.Credentials) error {
	if c.Login == "" {
		return fmt.Errorf("%w: login is empty", srvErrors.ErrInvalidCredentials)
	}

	if c.Password == "" {
		return fmt.Errorf("%w: password is empty", srvErrors.ErrInvalidCredentials)
	}

	if len(c.Login) > entity.UserMaxLoginLen {
		return fmt.Errorf(
			"%w: password too long, max %d characters",
			srvErrors.ErrInvalidCredentials,
			entity.UserMaxLoginLen,
		)
	}

	return nil
}
