package dto

import "time"

type Withdrawals struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawalsResp struct {
	Order     string    `json:"order"`
	Sum       float64   `json:"sum"`
	Processed time.Time `json:"processed_at"`
}
