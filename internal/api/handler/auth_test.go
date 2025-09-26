package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/mocks"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
)

func TestAuth_Register(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJleHAiOjE3NTg0NTk0OTMsImp0aSI6IjEifQ._mX-s6U9_iq4YhnQ5HOYbJAz7P8ly8BD_BufPYx2Kms"
	authHeader := "Bearer " + token

	errLoginTooLong := fmt.Errorf(
		"%w: password too long, max %d characters",
		errors.ErrAuthInvalidCredentials,
		entity.UserMaxLoginLen,
	)
	errPasswordTooLong := fmt.Errorf(
		"%w: password too long, max 72 bytes",
		errors.ErrAuthInvalidCredentials,
	)

	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		body  string
		setup func(t *testing.T) AuthService
		want  want
	}{
		{
			name: "success",
			body: `{"login":"testLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}).
					Return(token, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: authHeader,
				body:   "",
			},
		},
		{
			name: "negative_bad_json",
			body: `not valid jsson`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), gomock.All()).Times(0)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials format",
			},
		},
		{
			name: "negative_without_login",
			body: `{"password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Password: "t1estP5assword"}).
					Return("", errors.ErrAuthInvalidCredentials)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials",
			},
		},
		{
			name: "negative_without_password",
			body: `{"login":"testLogin"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "testLogin"}).
					Return("", errors.ErrAuthInvalidCredentials)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials",
			},
		},
		{
			name: "negative_login_too_long",
			body: `{"login":"veryLongLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "veryLongLogin", Password: "t1estP5assword"}).
					Return("", errLoginTooLong)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials: password too long, max 64 characters",
			},
		},
		{
			name: "negative_password_too_long",
			body: `{"login":"testLogin", "password":"veryLongPassword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "testLogin", Password: "veryLongPassword"}).
					Return("", errPasswordTooLong)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials: password too long, max 72 bytes",
			},
		},
		{
			name: "negative_user_already_exists",
			body: `{"login":"testLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}).
					Return("", errors.ErrAuthUserAlreadyExists)
				return service
			},
			want: want{
				code:   http.StatusConflict,
				header: "",
				body:   "user already exists",
			},
		},
		{
			name: "negative_server_error",
			body: `{"login":"testLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Register(gomock.All(), dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}).
					Return("", errors.ErrUnexpected)
				return service
			},
			want: want{
				code:   http.StatusInternalServerError,
				header: "",
				body:   statusText500,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			handler := NewAuth(service)

			reqBody := []byte(test.body)
			r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Register(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			resAuthHeader := res.Header.Get("Authorization")
			assert.Equal(t, test.want.header, resAuthHeader, "Response Authorization Header")
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}

func TestAuth_Login(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJleHAiOjE3NTg0NTk0OTMsImp0aSI6IjEifQ._mX-s6U9_iq4YhnQ5HOYbJAz7P8ly8BD_BufPYx2Kms"
	authHeader := "Bearer " + token

	type want struct {
		code   int
		header string
		body   string
	}

	tests := []struct {
		name  string
		body  string
		setup func(t *testing.T) AuthService
		want  want
	}{
		{
			name: "success",
			body: `{"login":"testLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Login(gomock.All(), dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}).
					Return(token, nil)
				return service
			},
			want: want{
				code:   http.StatusOK,
				header: authHeader,
				body:   "",
			},
		},
		{
			name: "negative_bad_json",
			body: `not valid jsson`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Login(gomock.All(), gomock.All()).Times(0)
				return service
			},
			want: want{
				code:   http.StatusBadRequest,
				header: "",
				body:   "invalid credentials format",
			},
		},
		{
			name: "negative_bad_login_or_password",
			body: `{"login":"badLogin", "password":"orPassword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Login(gomock.All(), dto.Credentials{Login: "badLogin", Password: "orPassword"}).
					Return("", errors.ErrAuthInvalidCredentials)
				return service
			},
			want: want{
				code:   http.StatusUnauthorized,
				header: "",
				body:   "",
			},
		},
		{
			name: "negative_server_error",
			body: `{"login":"testLogin", "password":"t1estP5assword"}`,
			setup: func(t *testing.T) AuthService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockAuthService(ctrl)
				service.EXPECT().
					Login(gomock.All(), dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}).
					Return("", errors.ErrUnexpected)
				return service
			},
			want: want{
				code:   http.StatusInternalServerError,
				header: "",
				body:   statusText500,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			handler := NewAuth(service)

			reqBody := []byte(test.body)
			r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Login(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			resAuthHeader := res.Header.Get("Authorization")
			assert.Equal(t, test.want.header, resAuthHeader, "Response Authorization Header")
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
