package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/dto"
)

const (
	pathPrefix       = "api/orders/"
	consumersTimeout = time.Duration(60 * time.Second)
)

var (
	ErrAccrualInternalError        = errors.New("internal server error")
	ErrAccrualUnexpectedStatusCode = errors.New("unexpected status code")
)

type ProcessingService interface {
	ProsessOrder(ctx context.Context, order dto.Order)
	MarkOrderForRetry(ctx context.Context, number string)
}

type Logger interface {
	Error(message string, err error)
	Warn(message string, err error)
}

type OrderConsumer struct {
	service ProcessingService
	logger  Logger
	client  *resty.Client
	address string
}

func NewOrderConsumer(srv ProcessingService, l Logger, serverAddr string) *OrderConsumer {
	serverAddr = strings.Trim(serverAddr, "/")

	return &OrderConsumer{
		service: srv,
		logger:  l,
		address: serverAddr,
		client: resty.New().
			SetTimeout(time.Duration(1) * time.Second).
			SetBaseURL(serverAddr + "/" + pathPrefix),
	}
}

func (c *OrderConsumer) Consume(ctx context.Context, number string) (retryAfter time.Duration) {
	var order dto.Order
	req := c.client.R().
		SetContext(ctx).
		SetResult(&order)

	resp, err := req.Get(number)
	if err != nil {
		c.logger.Warn("failed to request accrual servise", err)
		return
	}

	switch code := resp.StatusCode(); code {
	case http.StatusOK:
		if err := order.Validate(number); err == nil {
			c.service.ProsessOrder(ctx, order)
		} else {
			c.logger.Warn("invalid accrual service responce data", err)
			c.service.MarkOrderForRetry(ctx, number)
		}
	case http.StatusNoContent:
		c.service.MarkOrderForRetry(ctx, number)
	case http.StatusTooManyRequests:
		retryAfter = c.parseDelay(resp)
	case http.StatusInternalServerError:
		c.logger.Warn("accrual service internal error", ErrAccrualInternalError)
	default:
		c.logger.Error(
			"unexpected accrual service responce code",
			fmt.Errorf("%w: %d", ErrAccrualUnexpectedStatusCode, code),
		)
	}

	return
}

func (c *OrderConsumer) parseDelay(resp *resty.Response) time.Duration {
	delay := resp.Header().Get("Retry-After")

	v, err := strconv.ParseUint(delay, 10, 64)
	if err != nil {
		c.logger.Warn("failed to parse retry-after header", err)
		return consumersTimeout
	}

	return time.Duration(v) * time.Second
}
