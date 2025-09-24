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
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/mocks"
)

func TestWithdrawals_Withdraw(t *testing.T) {
	userID := uint64(13)
	userIDctx := context.WithValue(context.Background(), middleware.KeyUserID, userID)

	tests := []struct {
		name        string
		ctx         context.Context
		withdrawals dto.Withdrawals
		rSetup      func(t *testing.T) WithdrawalsRepository
		lSetup      func(t *testing.T) Logger
		want        error
	}{
		{
			name:        "success",
			ctx:         userIDctx,
			withdrawals: dto.Withdrawals{Order: "5062821234567892", Sum: 99.99},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), entity.Withdrawals{
						UserID:      userID,
						OrderNumber: "5062821234567892",
						Sum:         99.99,
					}).
					Return(nil)
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
			want: nil,
		},
		{
			name:        "negative_without_userID",
			ctx:         context.Background(),
			withdrawals: dto.Withdrawals{Order: "5062821234567892", Sum: 99.99},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
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
			want: srvErrors.ErrUnexpected,
		},
		{
			name:        "negative_ivalid_order_number",
			ctx:         userIDctx,
			withdrawals: dto.Withdrawals{Order: "5062821234567899", Sum: 99.99},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
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
			want: srvErrors.ErrOrderInvalidNumber,
		},
		{
			name:        "negative_ivalid_sum",
			ctx:         userIDctx,
			withdrawals: dto.Withdrawals{Order: "5062821234567892", Sum: 0.004},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
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
			want: srvErrors.ErrWithdrawInvalidSum,
		},
		{
			name:        "negative_insufficient_funds",
			ctx:         userIDctx,
			withdrawals: dto.Withdrawals{Order: "5062821234567892", Sum: 99.99},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), entity.Withdrawals{
						UserID:      userID,
						OrderNumber: "5062821234567892",
						Sum:         99.99,
					}).
					Return(repErrors.ErrNoRowsUpdated)
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
			want: srvErrors.ErrWithdrawInsufficientFunds,
		},
		{
			name:        "negative_repository_error",
			ctx:         userIDctx,
			withdrawals: dto.Withdrawals{Order: "5062821234567892", Sum: 99.99},
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), entity.Withdrawals{
						UserID:      userID,
						OrderNumber: "5062821234567892",
						Sum:         99.99,
					}).
					Return(fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to withdraw", gomock.All())
				return logger
			},
			want: srvErrors.ErrUnexpected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			service := NewWithdrawals(repository, logger)
			err := service.Withdraw(test.ctx, test.withdrawals)
			assert.ErrorIs(t, err, test.want, "Create withdrawals error")
		})
	}
}

func TestWithdrawals_List(t *testing.T) {
	userID := uint64(13)
	userIDctx := context.WithValue(context.Background(), middleware.KeyUserID, userID)

	entityList := []entity.Withdrawals{
		{OrderNumber: "5062821234567892", Sum: 99.99},
		{OrderNumber: "5062821234567819", Sum: 500},
	}
	dtoList := []dto.WithdrawalsResp{
		{Order: "5062821234567892", Sum: 99.99},
		{Order: "5062821234567819", Sum: 500},
	}

	type want struct {
		list []dto.WithdrawalsResp
		err  error
	}

	tests := []struct {
		name   string
		ctx    context.Context
		rSetup func(t *testing.T) WithdrawalsRepository
		lSetup func(t *testing.T) Logger
		want   want
	}{
		{
			name: "success",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(userIDctx, userID).
					Return(entityList, nil)
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
				list: dtoList,
				err:  nil,
			},
		},
		{
			name: "success_empty_list",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(userIDctx, userID).
					Return([]entity.Withdrawals{}, nil)
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
				list: []dto.WithdrawalsResp(nil),
				err:  nil,
			},
		},
		{
			name: "negative_without_userID",
			ctx:  context.Background(),
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(gomock.All(), gomock.All()).
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
				list: []dto.WithdrawalsResp(nil),
				err:  srvErrors.ErrUnexpected,
			},
		},
		{
			name: "negative_repository_error",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) WithdrawalsRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockWithdrawalsRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(userIDctx, userID).
					Return([]entity.Withdrawals{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to get user withdrawals", gomock.All())
				return logger
			},
			want: want{
				list: []dto.WithdrawalsResp(nil),
				err:  srvErrors.ErrUnexpected,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			service := NewWithdrawals(repository, logger)
			list, err := service.List(test.ctx)
			assert.Equal(t, test.want.list, list, "Get users withdrawals")
			assert.ErrorIs(t, err, test.want.err, "Get users withdrawals error")
		})
	}
}
