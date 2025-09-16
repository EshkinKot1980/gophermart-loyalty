package entity

import "time"

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number   string
	UserID   uint64
	Status   string
	Accrual  *float64
	Uploaded time.Time
	Updated  time.Time
}
