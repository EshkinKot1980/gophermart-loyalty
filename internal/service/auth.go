package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	FindByLogin(ctx context.Context, login string) (entity.User, error)
	GetByID(ctx context.Context, id uint64) (entity.User, error)
}

type Auth struct {
	repository UserRepository
	logger     Logger
	secret     string
}

func NewAuth(r UserRepository, l Logger, jwtSecret string) *Auth {
	return &Auth{repository: r, logger: l, secret: jwtSecret}
}

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
			err = fmt.Errorf("%w: password too long, max 72 bytes", srvErrors.ErrAuthInvalidCredentials)
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
			return token, srvErrors.ErrAuthUserAlreadyExists
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
			return token, srvErrors.ErrAuthInvalidCredentials
		} else {
			a.logger.Error("failed to find user", err)
			return token, srvErrors.ErrUnexpected
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(cr.Password))
	if err != nil {
		return token, srvErrors.ErrAuthInvalidCredentials
	}

	return a.generateToken(user)
}

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
		if !errors.Is(err, repErrors.ErrNotFound) {
			a.logger.Error("failed to find user by id", err)
		}
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

func (a *Auth) generateToken(u entity.User) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			ID:        strconv.FormatUint(u.ID, 10),
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
		return fmt.Errorf("%w: login is empty", srvErrors.ErrAuthInvalidCredentials)
	}

	if c.Password == "" {
		return fmt.Errorf("%w: password is empty", srvErrors.ErrAuthInvalidCredentials)
	}

	if len(c.Login) > entity.UserMaxLoginLen {
		return fmt.Errorf(
			"%w: password too long, max %d characters",
			srvErrors.ErrAuthInvalidCredentials,
			entity.UserMaxLoginLen,
		)
	}

	return nil
}
