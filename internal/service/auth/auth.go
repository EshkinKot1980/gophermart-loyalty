package auth

import (
	"context"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	FindByLogin(ctx context.Context, login string) (entity.User, error)
	GetByID(ctx context.Context, id uint64) (entity.User, error)
}

type Logger interface {
	Error(message string, err error)
}

type Auth struct {
	repository UserRepository
	logger     Logger
	secret     string
}

func New(r UserRepository, l Logger, jwtSecret string) *Auth {
	return &Auth{repository: r, logger: l, secret: jwtSecret}
}
