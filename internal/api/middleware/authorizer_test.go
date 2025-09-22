package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/middleware/mocks"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizer_Authorize(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJleHAiOjE3NTg0NTk0OTMsImp0aSI6IjEifQ._mX-s6U9_iq4YhnQ5HOYbJAz7P8ly8BD_BufPYx2Kms"
	authHeader := "Bearer " + token

	type want struct {
		code   int
		body   string
		userID uint64
	}

	tests := []struct {
		name   string
		header string
		setup  func(t *testing.T) AuthService
		want   want
	}{
		{
			name:   "success",
			header: authHeader,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					User(gomock.All(), token).
					Return(entity.User{ID: 13}, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				body:   "",
				userID: 13,
			},
		},
		{
			name:   "negative_without_header",
			header: "",
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					User(gomock.All(), "").
					Times(0)
				return service
			},
			want: want{
				code: http.StatusUnauthorized,
				body: "",
			},
		},
		{
			name:   "negative_token_expired",
			header: authHeader,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					User(gomock.All(), token).
					Return(entity.User{}, errors.ErrAuthTokenExpired)
				return service
			},
			want: want{
				code: http.StatusUnauthorized,
				body: "token expired",
			},
		},
		{
			name:   "negative_invalid_token",
			header: authHeader,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					User(gomock.All(), token).
					Return(entity.User{}, errors.ErrAuthInvalidToken)
				return service
			},
			want: want{
				code: http.StatusUnauthorized,
				body: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.header != "" {
				r.Header.Set("Authorization", test.header)
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(KeyUserID)
				assert.Equal(t, test.want.userID, userID, "Handler get userID")
				w.WriteHeader(http.StatusOK)
			})

			w := httptest.NewRecorder()
			mv := NewAuthorizer(test.setup(t))
			handler := mv.Authorize(next)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
