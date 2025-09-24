package processor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/processor/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderConsumer_Consume(t *testing.T) {
	type accrualResponse struct {
		netErr      bool
		status      int
		contentType string
		retryAfter  string
		body        string
	}

	tests := []struct {
		name     string
		number   string
		sleep    time.Duration
		sleeping bool
		resp     accrualResponse
		srvSetup func(t *testing.T) ProcessingService
		lSetup   func(t *testing.T) Logger
	}{
		{
			name:   "network_error",
			number: "5062821234567819",
			resp: accrualResponse{
				netErr: true,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Warn("failed to request accrual servise", gomock.All())
				return logger
			},
		},
		{
			name:   "network_error",
			number: "5062821234567819",
			resp: accrualResponse{
				netErr: true,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Warn("failed to request accrual servise", gomock.All())
				return logger
			},
		},
		{
			name:   "valid_response_data",
			number: "5062821234567892",
			resp: accrualResponse{
				status:      http.StatusOK,
				contentType: "application/json",
				body: `{
                            "order": "5062821234567892",
                            "status": "PROCESSED",
                            "accrual": 500
                        }`,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockProcessingService(ctrl)
				service.EXPECT().
					ProsessOrder(gomock.All(), dto.Order{
						Number:  "5062821234567892",
						Status:  dto.OrderStatusProcessed,
						Accrual: 500,
					})
				return service
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				return mocks.NewMockLogger(ctrl)
			},
		},
		{
			name:   "invalid_response_data",
			number: "5062821234567892",
			resp: accrualResponse{
				status:      http.StatusOK,
				contentType: "application/json",
				body: `{
                            "order": "5062821234567899",
                            "status": "PROCESSED",
                            "accrual": 500
                        }`,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockProcessingService(ctrl)
				service.EXPECT().
					MarkOrderForRetry(gomock.All(), "5062821234567892")
				return service
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Warn("invalid accrual service responce data", gomock.All())
				return logger
			},
		},
		{
			name:   "not_registred",
			number: "5062821234567819",
			resp: accrualResponse{
				status: http.StatusNoContent,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				service := mocks.NewMockProcessingService(ctrl)
				service.EXPECT().
					MarkOrderForRetry(gomock.All(), "5062821234567819")
				return service
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				return mocks.NewMockLogger(ctrl)
			},
		},
		{
			name:   "to_many_request",
			number: "5062821234567819",
			sleep:  13 * time.Second,
			resp: accrualResponse{
				status:     http.StatusTooManyRequests,
				retryAfter: "13",
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				return mocks.NewMockLogger(ctrl)
			},
		},
		{
			name:     "to_many_request_duplicate_signal",
			number:   "5062821234567819",
			sleep:    13 * time.Second,
			sleeping: true,
			resp: accrualResponse{
				status:     http.StatusTooManyRequests,
				retryAfter: "13",
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				return mocks.NewMockLogger(ctrl)
			},
		},
		{
			name:   "to_many_request_bad_header",
			number: "5062821234567819",
			sleep:  consumersTimeout,
			resp: accrualResponse{
				status:     http.StatusTooManyRequests,
				retryAfter: "bad_header",
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Warn("failed to parse retry-after header", gomock.All())
				return logger
			},
		},
		{
			name:   "accrual_servise_internal_error",
			number: "5062821234567819",
			sleep:  consumersTimeout,
			resp: accrualResponse{
				status: http.StatusInternalServerError,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Warn("accrual service internal error", gomock.All())
				return logger
			},
		},
		{
			name:   "unexpected_responce_staus_code",
			number: "5062821234567819",
			sleep:  consumersTimeout,
			resp: accrualResponse{
				status: http.StatusTeapot,
			},
			srvSetup: func(t *testing.T) ProcessingService {
				ctrl := gomock.NewController(t)
				return mocks.NewMockProcessingService(ctrl)
			},
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("unexpected accrual service responce code", gomock.All())
				return logger
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/"+pathPrefix+test.number, r.RequestURI, "Reqiest URI")
				assert.Equal(t, http.MethodGet, r.Method, "Request Method")

				w.Header().Set("Content-Type", test.resp.contentType)
				w.Header().Set("Retry-After", test.resp.retryAfter)
				w.WriteHeader(test.resp.status)
				_, err := w.Write([]byte(test.resp.body))
				require.Nil(t, err, "Write response body")
			}

			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			service := test.srvSetup(t)
			logger := test.lSetup(t)
			sleepAll := make(chan time.Duration, 1)
			defer close(sleepAll)

			consumer := NewOrderConsumer(service, logger, server.URL)
			if test.resp.netErr {
				server.Close()
			}
			if test.sleeping {
				sleepAll <- test.sleep
			}
			consumer.Consume(context.Background(), sleepAll, test.number)

			select {
			case sleep := <-sleepAll:
				assert.Equal(t, test.sleep, sleep, "Too many requests timeout")
			default:
			}
		})
	}
}
