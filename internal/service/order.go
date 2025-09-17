package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type OrderRepository interface {
	GetByNumber(ctx context.Context, number string) (entity.Order, error)
	GetAllByUser(ctx context.Context, userID uint64) ([]entity.Order, error)
	Create(ctx context.Context, order entity.Order) error
}

type Order struct {
	repository OrderRepository
	logger     Logger
}

func NewOrder(r OrderRepository, l Logger) *Order {
	return &Order{repository: r, logger: l}
}

func (o *Order) Upload(ctx context.Context, orderNumber string) error {
	userID, ok := ctx.Value(middleware.KeyUserID).(uint64)
	if !ok {
		o.logger.Error("failed to get user id", srvErrors.ErrUnexpected)
		return srvErrors.ErrUnexpected
	}

	if !isOrderNumberValid(orderNumber) {
		return srvErrors.ErrOrderInvalidNumber
	}

	newOrder := entity.Order{Number: orderNumber, UserID: userID}
	err := o.repository.Create(ctx, newOrder)
	if err == nil {
		return nil
	}

	if errors.Is(err, repErrors.ErrDuplicateKey) {
		return o.checkExistingOrder(ctx, orderNumber, userID)
	}
	o.logger.Error("failed to upload order", err)
	return srvErrors.ErrUnexpected
}

func (o *Order) List(ctx context.Context) ([]dto.Order, error) {
	var list []dto.Order

	userID, ok := ctx.Value(middleware.KeyUserID).(uint64)
	if !ok {
		o.logger.Error("failed to get user id", srvErrors.ErrUnexpected)
		return list, srvErrors.ErrUnexpected
	}

	orders, err := o.repository.GetAllByUser(ctx, userID)
	if err != nil {
		o.logger.Error("failed to get order user orders", err)
		return list, srvErrors.ErrUnexpected
	}

	for _, order := range orders {
		item := dto.Order{
			Number:   order.Number,
			Status:   order.Status,
			Uploaded: order.Updated,
		}
		if order.Accrual > 0 {
			item.Accrual = &order.Accrual
		}
		list = append(list, item)
	}

	return list, nil
}

func isOrderNumberValid(number string) bool {
	if _, err := strconv.ParseUint(number, 10, 64); err != nil {
		return false
	}

	sum := 0
	evenPos := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if evenPos {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		evenPos = !evenPos
	}

	return sum%10 == 0
}

func (o *Order) checkExistingOrder(ctx context.Context, orderNumber string, userID uint64) error {
	order, err := o.repository.GetByNumber(ctx, orderNumber)
	if err != nil {
		o.logger.Error("failed get existing order", err)
		return srvErrors.ErrUnexpected
	}
	if order.UserID == userID {
		return srvErrors.ErrOrderUploadedByUser
	}
	return srvErrors.ErrOrderUploadedByAnotherUser
}
