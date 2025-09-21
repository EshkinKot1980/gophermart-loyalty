package dto

import (
	"errors"
	"fmt"
)

var ErrInvalidResponseData = errors.New("invalid response data")

const (
	OrderStatusRegistred  = "REGISTERED"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (o Order) Validate(sentNumber string) error {
	if o.Number != sentNumber {
		return fmt.Errorf("%w: invalid order number", ErrInvalidResponseData)
	}

	switch o.Status {
	case OrderStatusRegistred:
	case OrderStatusProcessing:
	case OrderStatusInvalid:
	case OrderStatusProcessed:
		if o.Accrual < 0 {
			return fmt.Errorf("%w: invalid order accrual", ErrInvalidResponseData)
		}
	default:
		return fmt.Errorf("%w: invalid order staus:%s", ErrInvalidResponseData, o.Status)
	}

	return nil
}
