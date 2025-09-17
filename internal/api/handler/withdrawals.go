package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type WithdrawalsService interface {
	Withdraw(ctx context.Context, w dto.Withdrawals) error
	List(ctx context.Context) ([]dto.WithdrawalsResp, error)
}

type Withdrawals struct {
	service WithdrawalsService
	logger  Logger
}

func NewWithdrawals(srv WithdrawalsService, l Logger) *Withdrawals {
	return &Withdrawals{service: srv, logger: l}
}

func (h *Withdrawals) Withdraw(w http.ResponseWriter, r *http.Request) {
	var withdrawals dto.Withdrawals

	if err := json.NewDecoder(r.Body).Decode(&withdrawals); err != nil {
		http.Error(w, "invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.Withdraw(r.Context(), withdrawals)
	if err != nil {
		switch {
		case errors.Is(err, srvErrors.ErrInsufficientFunds):
			http.Error(w, "insufficient funds in the account", http.StatusPaymentRequired)
		case errors.Is(err, srvErrors.ErrWithdrawInvalidSum):
			http.Error(w, "sum must be positive", http.StatusUnprocessableEntity)
		case errors.Is(err, srvErrors.ErrOrderInvalidNumber):
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		default:
			http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Withdrawals) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	newJSONwriter(w, h.logger).write(list, "withdrawals list", http.StatusOK)
}
