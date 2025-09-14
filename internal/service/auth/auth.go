package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) error
	// GetByLogin(ctx context.Context, login string) (entity.User, error)
}

type Logger interface {
	Error(message string, err error)
}

type Auth struct {
	repository UserRepository
	logger     Logger
}

func New(r UserRepository, l Logger) *Auth {
	return &Auth{repository: r, logger: l}
}

func (a *Auth) Register(ctx context.Context, c dto.Credentials) error {
	cr := dto.Credentials{
		Login:    strings.TrimSpace(c.Login),
		Password: strings.TrimSpace(c.Password),
	}

	if err := validateCredentials(cr); err != nil {
		return err
	}

	// bcrypt имеет недостатки, в дальнейшем планирую переделать на другой алгоритм
	hash, err := bcrypt.GenerateFromPassword([]byte(cr.Password), bcrypt.DefaultCost)
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrPasswordTooLong):
			return fmt.Errorf("%w: password too long, max 72 bytes", srvErrors.ErrInvalidCredentials)
		default:
			a.logger.Error("failed to hash password", err)
			return srvErrors.ErrUnexpected
		}
	}

	user := entity.User{Login: c.Login, Hash: string(hash)}
	fmt.Println(user.Hash, len(user.Hash))
	if err := a.repository.Create(ctx, user); err != nil {
		switch {
		case errors.Is(err, repErrors.ErrDuplicateKey):
			return srvErrors.ErrUserAlreadyExists
		default:
			a.logger.Error("failed to create user", err)
			return srvErrors.ErrUnexpected
		}
	}

	return nil
}

func validateCredentials(c dto.Credentials) error {
	if c.Login == "" {
		return fmt.Errorf("%w: login is empty", srvErrors.ErrInvalidCredentials)
	}

	if len(c.Login) > entity.UserMaxLoginLen {
		return fmt.Errorf(
			"%w: password too long, max %d characters",
			srvErrors.ErrInvalidCredentials,
			entity.UserMaxLoginLen,
		)
	}

	if c.Password == "" {
		return fmt.Errorf("%w: password is empty", srvErrors.ErrInvalidCredentials)
	}

	return nil
}
