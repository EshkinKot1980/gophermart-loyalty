package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBalance_UserBalance(t *testing.T) {
	type want struct {
		code   int
		header string
		body   string
	}
	tests := []struct {
		name  string
		setup func(t *testing.T) BalanceService
		want  want
	}{
		{
			name: "success",
			setup: func(t *testing.T) BalanceService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockBalanceService(ctrl)
				service.EXPECT().
					UserBalance(gomock.All()).
					Return(dto.Balance{Current: 1310.8, Withdrawn: 800}, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: "application/json",
				body:   `{"current":1310.8,"withdrawn":800}`,
			},
		},
		{
			name: "negative",
			setup: func(t *testing.T) BalanceService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockBalanceService(ctrl)
				service.EXPECT().
					UserBalance(gomock.All()).
					Return(dto.Balance{}, errors.New("any error"))
				return service
			},
			want: want{
				code:   http.StatusInternalServerError,
				header: "text/plain",
				body:   "oops, something went wrong",
			},
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Error("", gomock.All()).Times(0)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			handler := NewBalance(service, logger)

			r := httptest.NewRequest(http.MethodGet, "/balance", nil)
			w := httptest.NewRecorder()
			handler.UserBalance(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			resContentType := w.Header().Get("Content-Type")
			assert.Contains(t, resContentType, test.want.header, "Response Content-Type")

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
