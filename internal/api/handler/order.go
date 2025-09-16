package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type OrderService interface {
	Upload(ctx context.Context, orderNumber string) error
}

type Order struct {
	service OrderService
}

func NewOrder(srv OrderService) *Order {
	return &Order{service: srv}
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if _, err := strconv.ParseUint(orderNumber, 10, 64); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	err = o.service.Upload(r.Context(), orderNumber)

	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, srvErrors.ErrOrderUploadedByUser):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, srvErrors.ErrOrderUploadedByAnotherUser):
		http.Error(w, "order already uploaded by another user", http.StatusConflict)
	case errors.Is(err, srvErrors.ErrOrderInvalidNumber):
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
	default:
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
	}
}
