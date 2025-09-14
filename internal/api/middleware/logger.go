package middleware

import (
	"net/http"
	"time"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
)

type Logger struct {
	logger HTTPloger
}

type requestData = logger.RequestLogData
type responseData = logger.ResponseLogData

type HTTPloger interface {
	RequestInfo(message string, req *requestData, resp *responseData)
}

func NewLogWraper(l HTTPloger) *Logger {
	return &Logger{logger: l}
}

func (l *Logger) Log(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			Status: 0,
			Size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		requestData := &requestData{
			URI:      r.RequestURI,
			Method:   r.Method,
			Duration: time.Since(start),
		}

		if responseData.Status == 0 {
			responseData.Status = http.StatusOK
		}

		l.logger.RequestInfo("server api", requestData, responseData)
	}

	return http.HandlerFunc(fn)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode
}
