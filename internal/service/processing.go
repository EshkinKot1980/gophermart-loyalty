package service

import (
	"context"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
)

type ProcessingRepository interface {
	OrderNubmersForProcess(ctx context.Context, statuses []string) ([]string, error)
	ProcessOrder(ctx context.Context, order entity.Order) error
	MarkOrderForRetryOrInvalid(
		ctx context.Context,
		number string,
		toStatus string,
		ivalidStatus string,
	) error
}

type Processing struct {
	reository ProcessingRepository
	logger    Logger
}

func NewProcessing(r ProcessingRepository, l Logger) *Processing {
	return &Processing{reository: r, logger: l}
}

func (p *Processing) ListForProccess(ctx context.Context) (orderNumbers []string) {
	statuses := []string{
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
	}

	orderNumbers, err := p.reository.OrderNubmersForProcess(ctx, statuses)
	if err != nil {
		p.logger.Error("failed to get orders for process", err)
		return orderNumbers
	}

	return orderNumbers
}

func (p *Processing) ProsessOrder(ctx context.Context, order dto.Order) {
	ent := entity.Order{
		Number:  order.Number,
		Status:  mapStatus(order.Status),
		Accrual: order.Accrual,
	}

	if ent.Status != entity.OrderStatusProcessed {
		ent.Accrual = 0
	}

	err := p.reository.ProcessOrder(ctx, ent)
	if err != nil {
		p.logger.Error("failed to process order", err)
	}
}

func (p *Processing) MarkOrderForRetry(ctx context.Context, number string) {
	err := p.reository.MarkOrderForRetryOrInvalid(
		ctx,
		number,
		entity.OrderStatusNew,
		entity.OrderStatusInvalid,
	)
	if err != nil {
		p.logger.Error("failed to mark order fo retry or invalid", err)
	}
}

func mapStatus(status string) string {
	switch status {
	case dto.OrderStatusRegistred, dto.OrderStatusProcessing:
		return entity.OrderStatusProcessing
	case dto.OrderStatusProcessed:
		return entity.OrderStatusProcessed
	case dto.OrderStatusInvalid:
		return entity.OrderStatusInvalid
	default:
		return status
	}
}
