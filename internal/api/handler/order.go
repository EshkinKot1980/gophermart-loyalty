package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

type OrderService interface {
	Upload(ctx context.Context, orderNumber string) error
	List(ctx context.Context) ([]dto.Order, error)
}

type Order struct {
	service OrderService
	logger  Logger
}

func NewOrder(srv OrderService, l Logger) *Order {
	return &Order{service: srv, logger: l}
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
		http.Error(w, statusText500, http.StatusInternalServerError)
	}
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	orders, err := o.service.List(r.Context())
	if err != nil {
		http.Error(w, statusText500, http.StatusInternalServerError)
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	newJSONwriter(w, o.logger).write(orders, "orders", http.StatusOK)
}
