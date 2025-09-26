package handler

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/dto"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/handler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_jsonWriter_write(t *testing.T) {
	type want struct {
		code   int
		header string
		body   string
	}

	validLSON := `{"number":"5062821234567892","status":"NEW","uploaded_at":"0001-01-01T00:00:00Z"}`

	tests := []struct {
		name       string
		witer      *testResponseWriter
		logger     func(t *testing.T) Logger
		value      any
		valueName  string
		stasusCode int
		want       want
	}{
		{
			name:  "positive",
			witer: newTestResponseWriter(false),
			logger: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error("", gomock.All()).Times(0)
				return logger
			},
			value: dto.Order{
				Number: "5062821234567892",
				Status: "NEW",
			},
			valueName:  "order",
			stasusCode: http.StatusOK,
			want: want{
				code:   http.StatusOK,
				header: "application/json",
				body:   validLSON,
			},
		},
		{
			name:  "negative_failed_to_json_encode",
			witer: newTestResponseWriter(false),
			logger: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error("failed to encode order to json", gomock.All()).
					Times(1)
				return logger
			},
			value:      math.Inf(-1),
			valueName:  "order",
			stasusCode: http.StatusOK,
			want: want{
				code:   http.StatusInternalServerError,
				header: "text/plain",
				body:   statusText500,
			},
		},
		{
			name:  "negative_failed_to_json_encode",
			witer: newTestResponseWriter(true),
			logger: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error("failed to write body", gomock.All()).
					Times(1)
				return logger
			},
			value: dto.Order{
				Number: "5062821234567892",
				Status: "NEW",
			},
			valueName:  "order",
			stasusCode: http.StatusOK,
			want: want{
				code:   http.StatusOK,
				header: "application/json",
				body:   "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := test.witer
			l := test.logger(t)

			jw := newJSONwriter(w, l)
			jw.write(test.value, test.valueName, test.stasusCode)

			assert.Equal(t, test.want.code, w.Code, "Response status code")
			assert.Contains(t, w.Header().Get("Content-Type"), test.want.header, "Response content type")
			resBody := strings.TrimSuffix(string(w.Body.String()), "\n")
			assert.Equal(t, test.want.body, resBody, "Response body")
		})
	}
}

// Нужен для того, чтобы протестировать ошибку в http.ResponseWriter.Write()
type testResponseWriter struct {
	httptest.ResponseRecorder
	needError bool
}

func newTestResponseWriter(needError bool) *testResponseWriter {
	recoder := &testResponseWriter{needError: needError}
	recoder.Body = new(bytes.Buffer)

	return recoder
}

func (trw *testResponseWriter) Write(buf []byte) (int, error) {
	if trw.needError {
		return 0, fmt.Errorf("unable to write, for test only")
	}
	return trw.Body.Write(buf)
}
