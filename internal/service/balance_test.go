package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/mocks"
)

func TestBalance_UserBalance(t *testing.T) {
	userID := uint64(13)
	userIDctx := context.WithValue(context.Background(), middleware.KeyUserID, userID)

	type want struct {
		balance dto.Balance
		err     error
	}

	tests := []struct {
		name   string
		ctx    context.Context
		rSetup func(t *testing.T) BalanceRepository
		lSetup func(t *testing.T) Logger
		want   want
	}{
		{
			name: "success",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) BalanceRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockBalanceRepository(ctrl)
				repository.EXPECT().
					GetByUser(gomock.All(), userID).
					Return(entity.Balance{Balance: 599.99, Debited: 400}, nil)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				balance: dto.Balance{Current: 599.99, Withdrawn: 400},
				err:     nil,
			},
		},
		{
			name: "negative_without_userID",
			ctx:  context.Background(),
			rSetup: func(t *testing.T) BalanceRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockBalanceRepository(ctrl)
				repository.EXPECT().
					GetByUser(gomock.All(), userID).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to get user id", gomock.All())
				return logger
			},
			want: want{
				balance: dto.Balance{},
				err:     errors.ErrUnexpected,
			},
		},
		{
			name: "negative_repository_error",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) BalanceRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockBalanceRepository(ctrl)
				repository.EXPECT().
					GetByUser(gomock.All(), userID).
					Return(entity.Balance{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to get user balance", gomock.All())
				return logger
			},
			want: want{
				balance: dto.Balance{},
				err:     errors.ErrUnexpected,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)

			balanceService := NewBalance(repository, logger)
			balance, err := balanceService.UserBalance(test.ctx)

			assert.Equal(t, test.want.balance, balance, "Get user balance dto")
			assert.ErrorIs(t, err, test.want.err, "Get user balance error")
		})
	}
}
