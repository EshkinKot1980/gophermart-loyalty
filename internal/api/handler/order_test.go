package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/mocks"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testUnreadebleBody struct{}

func (tub *testUnreadebleBody) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("unable to read, for test only")
}

func TestOrder_Create(t *testing.T) {
	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		body  string
		setup func(t *testing.T) OrderService
		want  want
	}{
		{
			name: "success_uploaded",
			body: "5062821234567892",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					Upload(gomock.All(), "5062821234567892").
					Return(nil)
				return service
			},
			want: want{
				code:   http.StatusAccepted,
				header: "",
				body:   "",
			},
		},
		{
			name: "success_already_uploaded",
			body: "5062821234567892",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					Upload(gomock.All(), "5062821234567892").
					Return(errors.ErrOrderUploadedByUser)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: "",
				body:   "",
			},
		},
		{
			name: "negative_failed_to_read_body",
			body: "",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().Upload(gomock.All(), "").Times(0)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "text/plain",
				body:   "failed to read body",
			},
		},
		{
			name: "negative_invalid_format",
			body: "50628212 34567892",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().Upload(gomock.All(), "").Times(0)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "text/plain",
				body:   "invalid request format",
			},
		},
		{
			name: "negative_uploaded_by_another_user",
			body: "5062821234567892",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					Upload(gomock.All(), "5062821234567892").
					Return(errors.ErrOrderUploadedByAnotherUser)
				return service
			},
			want: want{
				code:   http.StatusConflict,
				header: "text/plain",
				body:   "order already uploaded by another user",
			},
		},
		{
			name: "negative_invalid_order_number",
			body: "5062821234567899",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					Upload(gomock.All(), "5062821234567899").
					Return(errors.ErrOrderInvalidNumber)
				return service
			},
			want: want{
				code:   http.StatusUnprocessableEntity,
				header: "text/plain",
				body:   "invalid order number",
			},
		},
		{
			name: "negative_server_error",
			body: "5062821234567892",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					Upload(gomock.All(), "5062821234567892").
					Return(errors.ErrUnexpected)
				return service
			},
			want: want{
				code:   http.StatusInternalServerError,
				header: "text/plain",
				body:   statusText500,
			},
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Error("", gomock.All()).Times(0)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			handler := NewOrder(service, logger)

			var respBody io.Reader
			if test.body != "" {
				body := []byte(test.body)
				respBody = bytes.NewBuffer(body)
			} else {
				respBody = &testUnreadebleBody{}
			}

			r := httptest.NewRequest(http.MethodPost, "/orders", respBody)
			w := httptest.NewRecorder()
			handler.Create(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			if test.want.header != "" {
				resContentType := w.Header().Get("Content-Type")
				assert.Contains(t, resContentType, test.want.header, "Response Content-Type")
			}

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}

func TestOrder_List(t *testing.T) {
	accrual := 99.99
	orders := []dto.Order{
		{
			Number: "5062821234567892",
			Status: entity.OrderStatusInvalid,
		},
		{
			Number:  "5062821234567819",
			Status:  entity.OrderStatusProcessed,
			Accrual: &accrual,
		},
	}

	successBody := `[` +
		`{"number":"5062821234567892","status":"INVALID","uploaded_at":"0001-01-01T00:00:00Z"},` +
		`{"number":"5062821234567819","status":"PROCESSED","accrual":99.99,` +
		`"uploaded_at":"0001-01-01T00:00:00Z"}` +
		`]`

	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		setup func(t *testing.T) OrderService
		want  want
	}{
		{
			name: "success",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return(orders, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: "application/json",
				body:   successBody,
			},
		},
		{
			name: "success_empty_list",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return([]dto.Order{}, nil)
				return service
			},
			want: want{
				code:   http.StatusNoContent,
				header: "",
				body:   "",
			},
		},
		{
			name: "negative_server_error",
			setup: func(t *testing.T) OrderService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockOrderService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return([]dto.Order{}, errors.ErrUnexpected)
				return service
			},
			want: want{
				code:   http.StatusInternalServerError,
				header: "text/plain",
				body:   statusText500,
			},
		},
	}
	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Error("", gomock.All()).Times(0)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			handler := NewOrder(service, logger)

			r := httptest.NewRequest(http.MethodGet, "/orders", nil)
			w := httptest.NewRecorder()
			handler.List(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			if test.want.header != "" {
				resContentType := w.Header().Get("Content-Type")
				assert.Contains(t, resContentType, test.want.header, "Response Content-Type")
			}

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
