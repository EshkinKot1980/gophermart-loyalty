package service

import (
	"context"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type BalanceRepository interface {
	GetByUser(ctx context.Context, userID uint64) (entity.Balance, error)
}

type Balance struct {
	repository BalanceRepository
	logger     Logger
}

func NewBalance(r BalanceRepository, l Logger) *Balance {
	return &Balance{repository: r, logger: l}
}

func (b *Balance) UserBalance(ctx context.Context) (balance dto.Balance, err error) {
	userID, ok := ctx.Value(middleware.KeyUserID).(uint64)
	if !ok {
		b.logger.Error("failed to get user id", errors.ErrUnexpected)
		return balance, errors.ErrUnexpected
	}

	entity, err := b.repository.GetByUser(ctx, userID)
	if err != nil {
		b.logger.Error("failed to get user balance", err)
		return balance, errors.ErrUnexpected
	}

	balance.Current = entity.Balance
	balance.Withdrawn = entity.Debited

	return balance, nil
}
