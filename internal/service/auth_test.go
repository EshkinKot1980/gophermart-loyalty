package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/entity"
	repErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/repository/errors"
	srvErrors "github.com/EshkinKot1980/gophermart-loyalty/internal/service/errors"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service/mocks"
)

func TestAuth_Register(t *testing.T) {
	jwtSecret := "secret"

	goodCredentials := dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}

	tooLongLoginCr := dto.Credentials{Login: "l", Password: "t1estP5assword"}
	for range entity.UserMaxLoginLen {
		tooLongLoginCr.Login += "l"
	}

	tooLongPasswordCr := dto.Credentials{Login: "testLogin", Password: "p"}
	for range 72 {
		tooLongPasswordCr.Password += "p"
	}

	type want struct {
		userID uint64
		err    error
	}

	tests := []struct {
		name        string
		credentials dto.Credentials
		rSetup      func(t *testing.T) UserRepository
		lSetup      func(t *testing.T) Logger
		want        want
	}{
		{
			name:        "succes",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Return(entity.User{ID: 13}, nil)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				userID: 13,
				err:    nil,
			},
		},
		{
			name:        "negative_empty_login",
			credentials: dto.Credentials{Login: "", Password: "t1estP5assword"},
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_empty_password",
			credentials: dto.Credentials{Login: "testLogin", Password: ""},
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_too_long_login",
			credentials: tooLongLoginCr,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_too_long_password",
			credentials: tooLongPasswordCr,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_user_exist",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Return(entity.User{}, repErrors.ErrDuplicateKey)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthUserAlreadyExists,
			},
		},
		{
			name:        "negative_unexpected_repository_error",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					Create(gomock.All(), gomock.All()).
					Return(entity.User{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to create user", gomock.All())
				return logger
			},
			want: want{
				err: srvErrors.ErrUnexpected,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			ctx := context.Background()

			authService := NewAuth(repository, logger, jwtSecret)
			token, err := authService.Register(ctx, test.credentials)

			assert.ErrorIs(t, err, test.want.err, "Register user error")
			if err != nil {
				return
			}

			jt, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})
			require.Nil(t, err, "Parse token")
			userID, err := parseID(jt)
			require.Nil(t, err, "Get ID from token")
			assert.Equal(t, test.want.userID, userID, "Registered userID form token")
		})
	}
}

func TestAuth_Login(t *testing.T) {
	jwtSecret := "secret"

	goodCredentials := dto.Credentials{Login: "testLogin", Password: "t1estP5assword"}
	hash, err := bcrypt.GenerateFromPassword([]byte("t1estP5assword"), bcrypt.DefaultCost)
	require.Nil(t, err, "Generate hash for entity")

	badPasswordCredentials := dto.Credentials{Login: "testLogin", Password: "badPassword"}

	type want struct {
		userID uint64
		err    error
	}

	tests := []struct {
		name        string
		credentials dto.Credentials
		rSetup      func(t *testing.T) UserRepository
		lSetup      func(t *testing.T) Logger
		want        want
	}{
		{
			name:        "succes",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					FindByLogin(gomock.All(), goodCredentials.Login).
					Return(entity.User{ID: 13, Hash: string(hash)}, nil)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				userID: 13,
				err:    nil,
			},
		},
		{
			name:        "negative_user_not_found",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					FindByLogin(gomock.All(), goodCredentials.Login).
					Return(entity.User{}, repErrors.ErrNotFound)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_bad_password",
			credentials: badPasswordCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					FindByLogin(gomock.All(), badPasswordCredentials.Login).
					Return(entity.User{ID: 13, Hash: string(hash)}, nil)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				err: srvErrors.ErrAuthInvalidCredentials,
			},
		},
		{
			name:        "negative_unexpected_repository_error",
			credentials: goodCredentials,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					FindByLogin(gomock.All(), goodCredentials.Login).
					Return(entity.User{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to find user", gomock.All())
				return logger
			},
			want: want{
				err: srvErrors.ErrUnexpected,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			ctx := context.Background()

			authService := NewAuth(repository, logger, jwtSecret)
			token, err := authService.Login(ctx, test.credentials)

			assert.ErrorIs(t, err, test.want.err, "Login user error")
			if err != nil {
				return
			}

			jt, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})
			require.Nil(t, err, "Parse token")
			userID, err := parseID(jt)
			require.Nil(t, err, "Get ID from token")
			assert.Equal(t, test.want.userID, userID, "Logged in userID form token")
		})
	}
}

func TestAuth_User(t *testing.T) {
	jwtSecret := "secret"
	goodToken := testGenerateToken(t, 13, jwtSecret, false)
	expiredToken := testGenerateToken(t, 13, jwtSecret, true)
	badSignedToken := testGenerateToken(t, 13, "bad secret", false)
	badIDtoken := testGenerateToken(t, 0, jwtSecret, false)

	type want struct {
		user entity.User
		err  error
	}

	tests := []struct {
		name   string
		token  string
		rSetup func(t *testing.T) UserRepository
		lSetup func(t *testing.T) Logger
		want   want
	}{
		{
			name:  "success",
			token: goodToken,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					GetByID(gomock.All(), uint64(13)).
					Return(entity.User{ID: 13}, nil)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				user: entity.User{ID: 13},
				err:  nil,
			},
		},
		{
			name:  "negative_token_expired",
			token: expiredToken,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					GetByID(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				user: entity.User{},
				err:  srvErrors.ErrAuthTokenExpired,
			},
		},
		{
			name:  "negative_bad_signed_token",
			token: badSignedToken,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					GetByID(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				user: entity.User{},
				err:  srvErrors.ErrAuthInvalidToken,
			},
		},
		{
			name:  "negative_token_without_id",
			token: badIDtoken,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					GetByID(gomock.All(), gomock.All()).
					Times(0)
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("", gomock.All()).
					Times(0)
				return logger
			},
			want: want{
				user: entity.User{},
				err:  srvErrors.ErrAuthInvalidToken,
			},
		},
		{
			name:  "negative_unexpected_repository_error",
			token: goodToken,
			rSetup: func(t *testing.T) UserRepository {
				ctrl := gomock.NewController(t)
				repository := mocks.NewMockUserRepository(ctrl)
				repository.EXPECT().
					GetByID(gomock.All(), uint64(13)).
					Return(entity.User{}, fmt.Errorf("any error"))
				return repository
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to find user by id", gomock.All())
				return logger
			},
			want: want{
				user: entity.User{},
				err:  srvErrors.ErrAuthInvalidToken,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.rSetup(t)
			logger := test.lSetup(t)
			ctx := context.Background()

			authService := NewAuth(repository, logger, jwtSecret)
			user, err := authService.User(ctx, test.token)

			assert.Equal(t, test.want.user, user, "Get user entity")
			assert.ErrorIs(t, err, test.want.err, "Get user error")
		})
	}
}

func testGenerateToken(t *testing.T, id uint64, secret string, expired bool) string {
	expires := time.Now().Add((-1) * time.Hour)
	if !expired {
		expires = expires.Add(25 * time.Hour)
	}

	idStr := strconv.FormatUint(id, 10)
	if id == 0 {
		idStr = ""
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			ID:        idStr,
		},
	)

	tokenStr, err := token.SignedString([]byte(secret))
	require.Nil(t, err, "Token generation")

	return tokenStr
}
