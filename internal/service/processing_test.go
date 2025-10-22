package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/mocks"
)

func TestProcessing_ListForProccess(t *testing.T) {
	orderNumbers := []string{"5062821234567892", "5062821234567819"}
	statuses := []string{
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
	}

	tests := []struct {
		name   string
		rSetup func(t *testing.T) ProcessingRepository
		lSetup func(t *testing.T) Logger
		want   []string
	}{
		{
			name: "success",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					OrderNubmersForProcess(gomock.All(), statuses).
					Return(orderNumbers, nil)
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
			want: orderNumbers,
		},
		{
			name: "empty_list",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					OrderNubmersForProcess(gomock.All(), statuses).
					Return([]string{}, nil)
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
			want: []string{},
		},
		{
			name: "repository_error",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					OrderNubmersForProcess(gomock.All(), statuses).
					Return([]string{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to get orders for process", gomock.All())
				return logger
			},
			want: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			processingService := NewProcessing(repository, logger)
			list := processingService.ListForProccess(context.Background())
			assert.Equal(t, test.want, list, "Get orders numbers")
		})
	}
}

func TestProcessing_ProsessOrder(t *testing.T) {
	orderDTO := dto.Order{
		Number:  "5062821234567892",
		Status:  dto.OrderStatusRegistred,
		Accrual: 100,
	}
	orderEntiy := entity.Order{
		Number: "5062821234567892",
		Status: entity.OrderStatusProcessing,
	}

	tests := []struct {
		name   string
		rSetup func(t *testing.T) ProcessingRepository
		lSetup func(t *testing.T) Logger
	}{
		{
			name: "success",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					ProcessOrder(gomock.All(), orderEntiy).
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
		},
		{
			name: "repository_error",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					ProcessOrder(gomock.All(), orderEntiy).
					Return(fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to process order", gomock.All())
				return logger
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			processingService := NewProcessing(repository, logger)
			processingService.ProsessOrder(context.Background(), orderDTO)
		})
	}
}

func TestProcessing_MarkOrderForRetry(t *testing.T) {
	orderNumber := "5062821234567892"

	tests := []struct {
		name   string
		rSetup func(t *testing.T) ProcessingRepository
		lSetup func(t *testing.T) Logger
	}{
		{
			name: "success",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					MarkOrderForRetryOrInvalid(
						gomock.All(),
						orderNumber,
						entity.OrderStatusNew,
						entity.OrderStatusInvalid,
					).
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
		},
		{
			name: "repository_error",
			rSetup: func(t *testing.T) ProcessingRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockProcessingRepository(ctrl)
				repository.EXPECT().
					MarkOrderForRetryOrInvalid(
						gomock.All(),
						orderNumber,
						entity.OrderStatusNew,
						entity.OrderStatusInvalid,
					).
					Return(fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to mark order fo retry or invalid", gomock.All())
				return logger
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			processingService := NewProcessing(repository, logger)
			processingService.MarkOrderForRetry(context.Background(), orderNumber)
		})
	}
}

func Test_mapStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
	}{
		{
			name:   "registred",
			status: dto.OrderStatusRegistred,
			want:   entity.OrderStatusProcessing,
		},
		{
			name:   "processing",
			status: dto.OrderStatusProcessing,
			want:   entity.OrderStatusProcessing,
		},
		{
			name:   "processed",
			status: dto.OrderStatusProcessed,
			want:   entity.OrderStatusProcessed,
		},
		{
			name:   "invalid",
			status: dto.OrderStatusInvalid,
			want:   entity.OrderStatusInvalid,
		},
		{
			name:   "any_other_status",
			status: "ANY_OTHER",
			want:   "ANY_OTHER",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mapStatus(test.status)
			if test.want != got {
				t.Errorf("mapStatus() = %v, want %v", got, test.want)
			}
		})
	}
}
