package handler

import (
	"context"
	"net/http"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
)

type BalanceService interface {
	UserBalance(ctx context.Context) (dto.Balance, error)
}

type Balance struct {
	service BalanceService
	logger  Logger
}

func NewBalance(srv BalanceService, l Logger) *Balance {
	return &Balance{service: srv, logger: l}
}

func (b *Balance) UserBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := b.service.UserBalance(r.Context())
	if err != nil {
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
	}

	newJSONwriter(w, b.logger).write(balance, "balance", http.StatusOK)
}
