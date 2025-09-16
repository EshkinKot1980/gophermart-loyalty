package entity

import "time"

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number   string    `db:"number"`
	UserID   uint64    `db:"user_id"`
	Status   string    `db:"status"`
	Accrual  float64   `db:"accrual"`
	Uploaded time.Time `db:"uploaded_at"`
	Updated  time.Time `db:"updated_at"`
}
