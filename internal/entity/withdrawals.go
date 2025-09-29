package entity

import "time"

type Withdrawals struct {
	ID          uint64    `db:"id"`
	UserID      uint64    `db:"user_id"`
	OrderNumber string    `db:"order_num"`
	Sum         float64   `db:"sum"`
	Processed   time.Time `db:"processed_at"`
}
