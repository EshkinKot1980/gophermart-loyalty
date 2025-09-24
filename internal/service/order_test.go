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

func TestOrder_Upload(t *testing.T) {
	userID := uint64(13)
	userIDctx := context.WithValue(context.Background(), middleware.KeyUserID, userID)
	orderToCreate := entity.Order{Number: "5062821234567892", UserID: userID}

	tests := []struct {
		name   string
		ctx    context.Context
		number string
		rSetup func(t *testing.T) OrderRepository
		lSetup func(t *testing.T) Logger
		want   error
	}{
		{
			name:   "success",
			number: "5062821234567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), orderToCreate).
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
			name:   "negative_without_userID",
			number: "5062821234567892",
			ctx:    context.Background(),
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
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
			name:   "negative_ivalid_order_number",
			number: "50628212 34567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
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
			name:   "success_uploaded_by_user",
			number: "5062821234567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), orderToCreate).
					Return(repErrors.ErrDuplicateKey)
				repository.EXPECT().
					GetByNumber(gomock.All(), "5062821234567892").
					Return(entity.Order{UserID: userID}, nil)
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
			want: srvErrors.ErrOrderUploadedByUser,
		},
		{
			name:   "negative_uploaded_by_another_user",
			number: "5062821234567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), orderToCreate).
					Return(repErrors.ErrDuplicateKey)
				repository.EXPECT().
					GetByNumber(gomock.All(), "5062821234567892").
					Return(entity.Order{UserID: 10}, nil)
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
			want: srvErrors.ErrOrderUploadedByAnotherUser,
		},
		{
			name:   "negative_repository_get_error",
			number: "5062821234567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), orderToCreate).
					Return(repErrors.ErrDuplicateKey)
				repository.EXPECT().
					GetByNumber(gomock.All(), "5062821234567892").
					Return(entity.Order{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed get existing order", gomock.All())
				return logger
			},
			want: srvErrors.ErrUnexpected,
		},
		{
			name:   "negative_repository_create_error",
			number: "5062821234567892",
			ctx:    userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), orderToCreate).
					Return(fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to upload order", gomock.All())
				return logger
			},
			want: srvErrors.ErrUnexpected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			orderService := NewOrder(repository, logger)
			err := orderService.Upload(test.ctx, test.number)
			assert.ErrorIs(t, err, test.want, "Create order error")
		})
	}
}

func TestOrder_List(t *testing.T) {
	userID := uint64(13)
	userIDctx := context.WithValue(context.Background(), middleware.KeyUserID, userID)
	orderEntities := []entity.Order{
		{Number: "5062821234567892", Status: entity.OrderStatusNew},
		{Number: "5062821234567819", Status: entity.OrderStatusProcessed, Accrual: 100},
	}
	orderDTOlist := []dto.Order{
		{Number: "5062821234567892", Status: entity.OrderStatusNew},
		{
			Number:  "5062821234567819",
			Status:  entity.OrderStatusProcessed,
			Accrual: &orderEntities[1].Accrual,
		},
	}

	type want struct {
		list []dto.Order
		err  error
	}

	tests := []struct {
		name   string
		ctx    context.Context
		rSetup func(t *testing.T) OrderRepository
		lSetup func(t *testing.T) Logger
		want   want
	}{
		{
			name: "success",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(gomock.All(), userID).
					Return(orderEntities, nil)
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
				list: orderDTOlist,
				err:  nil,
			},
		},
		{
			name: "success_empty_list",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(gomock.All(), userID).
					Return([]entity.Order{}, nil)
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
				list: []dto.Order(nil),
				err:  nil,
			},
		},
		{
			name: "negative_without_userID",
			ctx:  context.Background(),
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(gomock.All(), userID).
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
				list: []dto.Order(nil),
				err:  srvErrors.ErrUnexpected,
			},
		},
		{
			name: "negative_repository_error",
			ctx:  userIDctx,
			rSetup: func(t *testing.T) OrderRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockOrderRepository(ctrl)
				repository.EXPECT().
					GetAllByUser(gomock.All(), userID).
					Return([]entity.Order{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to get user orders", gomock.All())
				return logger
			},
			want: want{
				list: []dto.Order(nil),
				err:  srvErrors.ErrUnexpected,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			orderService := NewOrder(repository, logger)
			list, err := orderService.List(test.ctx)
			assert.Equal(t, test.want.list, list, "Get user orders error")
			assert.ErrorIs(t, err, test.want.err, "Get user orders error")
		})
	}
}
