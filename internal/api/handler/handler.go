package handler

import (
	"encoding/json"
	"net/http"
)

type Logger interface {
	Error(message string, err error)
}

type jsonWriter struct {
	writer http.ResponseWriter
	logger Logger
}

func newJSONwriter(w http.ResponseWriter, l Logger) *jsonWriter {
	return &jsonWriter{writer: w, logger: l}
}

func (jw *jsonWriter) write(value any, valueName string, stasusCode int) {
	body, err := json.Marshal(value)
	if err != nil {
		jw.logger.Error("failed to encode "+valueName+" to json", err)
		http.Error(jw.writer, "oops, something went wrong", http.StatusInternalServerError)
		return
	}

	jw.writer.Header().Set("Content-Type", "application/json")
	jw.writer.WriteHeader(stasusCode)

	_, err = jw.writer.Write([]byte(body))
	if err != nil {
		jw.logger.Error("failed to write body", err)
	}
}
