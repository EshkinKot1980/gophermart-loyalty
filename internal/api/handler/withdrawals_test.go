package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/mocks"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

func TestWithdrawals_Withdraw(t *testing.T) {
	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		body  string
		setup func(t *testing.T) WithdrawalsService
		want  want
	}{
		{
			name: "success",
			body: `{"order":"5062821234567892", "sum":700}`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{Order: "5062821234567892", Sum: 700}).
					Return(nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: "",
				body:   "",
			},
		},
		{
			name: "negative_invalid_format",
			body: `not valid jsson`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{}).
					Times(0)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "text/plain",
				body:   "invalid request format",
			},
		},
		{
			name: "negative_insufficient_funds",
			body: `{"order":"5062821234567892", "sum":700}`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{Order: "5062821234567892", Sum: 700}).
					Return(errors.ErrWithdrawInsufficientFunds)
				return service
			},
			want: want{
				code:   http.StatusPaymentRequired,
				header: "text/plain",
				body:   "insufficient funds in the account",
			},
		},
		{
			name: "negative_invalid_sum",
			body: `{"order":"5062821234567892", "sum":0}`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{Order: "5062821234567892", Sum: 0}).
					Return(errors.ErrWithdrawInvalidSum)
				return service
			},
			want: want{
				code:   http.StatusUnprocessableEntity,
				header: "text/plain",
				body:   "sum must be positive",
			},
		},
		{
			name: "negative_invalid_order",
			body: `{"order":"5062821234567899", "sum":700}`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{Order: "5062821234567899", Sum: 700}).
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
			body: `{"order":"5062821234567892", "sum":700}`,
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					Withdraw(gomock.All(), dto.Withdrawals{Order: "5062821234567892", Sum: 700}).
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
			handler := NewWithdrawals(service, logger)

			respBody := []byte(test.body)
			r := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(respBody))
			w := httptest.NewRecorder()
			handler.Withdraw(w, r)
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

func TestWithdrawals_List(t *testing.T) {
	list := []dto.WithdrawalsResp{{Order: "5062821234567892", Sum: 700}}
	listBody := `[{"order":"5062821234567892","sum":700,"processed_at":"0001-01-01T00:00:00Z"}]`

	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		setup func(t *testing.T) WithdrawalsService
		want  want
	}{
		{
			name: "success",
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return(list, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: "application/json",
				body:   listBody,
			},
		},
		{
			name: "success_empty_list",
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return([]dto.WithdrawalsResp{}, nil)
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
			setup: func(t *testing.T) WithdrawalsService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockWithdrawalsService(ctrl)
				service.EXPECT().
					List(gomock.All()).
					Return([]dto.WithdrawalsResp{}, errors.ErrUnexpected)
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
			handler := NewWithdrawals(service, logger)

			r := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
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
