package middleware

import (
	// "bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	lm := makeLoggerMock(t)
	mw := NewLogWraper(lm)
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "four")
	})

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	handler := mw.Log(next)
	handler.ServeHTTP(w, r)
}

type httpLoggerMock struct {
	t *testing.T
}

func makeLoggerMock(t *testing.T) httpLoggerMock {
	return httpLoggerMock{t: t}
}

func (m httpLoggerMock) RequestInfo(message string, req *requestData, resp *responseData) {
	assert.Equal(m.t, "server api", message, "Log message")
	assert.Equal(m.t, "/path", req.URI, "Log request URI")
	assert.Equal(m.t, http.MethodGet, req.Method, "Log request method")
	assert.NotEmpty(m.t, req.Duration, "Log request duration")
	assert.Equal(m.t, http.StatusNotFound, resp.Status, "Log response status code")
	assert.Equal(m.t, 4, resp.Size, "Log response content length")
}
