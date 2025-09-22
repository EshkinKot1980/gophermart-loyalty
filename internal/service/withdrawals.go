package service

import (
	"context"
	"errors"
	"math"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type WithdrawalsRepository interface {
	Create(ctx context.Context, widrawals entity.Withdrawals) error
	GetAllByUser(ctx context.Context, userID uint64) ([]entity.Withdrawals, error)
}

type Withdrawals struct {
	repository WithdrawalsRepository
	logger     Logger
}

func NewWithdrawals(r WithdrawalsRepository, l Logger) *Withdrawals {
	return &Withdrawals{repository: r, logger: l}
}

func (s *Withdrawals) Withdraw(ctx context.Context, w dto.Withdrawals) error {
	userID, ok := ctx.Value(middleware.KeyUserID).(uint64)
	if !ok {
		s.logger.Error("failed to get user id", srvErrors.ErrUnexpected)
		return srvErrors.ErrUnexpected
	}

	if !isOrderNumberValid(w.Order) {
		return srvErrors.ErrOrderInvalidNumber
	}
	if math.Round(w.Sum*100)/100 <= 0 {
		return srvErrors.ErrWithdrawInvalidSum
	}

	entity := entity.Withdrawals{UserID: userID, OrderNumber: w.Order, Sum: w.Sum}
	err := s.repository.Create(ctx, entity)
	if err == nil {
		return nil
	}

	if errors.Is(err, repErrors.ErrNoRowsUpdated) {
		return srvErrors.ErrWithdrawInsufficientFunds
	}
	s.logger.Error("failed to withdraw", err)
	return srvErrors.ErrUnexpected
}

func (s *Withdrawals) List(ctx context.Context) (list []dto.WithdrawalsResp, err error) {
	userID, ok := ctx.Value(middleware.KeyUserID).(uint64)
	if !ok {
		s.logger.Error("failed to get user id", srvErrors.ErrUnexpected)
		return list, srvErrors.ErrUnexpected
	}

	entities, err := s.repository.GetAllByUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user withdrawals", err)
		return list, srvErrors.ErrUnexpected
	}

	for _, entity := range entities {
		list = append(list, dto.WithdrawalsResp{
			Order:     entity.OrderNumber,
			Sum:       entity.Sum,
			Processed: entity.Processed,
		})
	}

	return list, nil
}
