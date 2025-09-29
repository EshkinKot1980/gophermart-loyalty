package entity

type Balance struct {
	UserID  uint64  `db:"user_id"`
	Balance float64 `db:"balance"`
	Debited float64 `db:"debited"`
}
