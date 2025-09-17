package handler

import (
	"encoding/json"
	"net/http"
)

type Logger interface {
	Error(message string, err error)
}

func writeJSON(v any, vName string, stasusCode int, w http.ResponseWriter, l Logger) {
	body, err := json.Marshal(v)
	if err != nil {
		l.Error("failed to encode "+vName+" to json", err)
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stasusCode)
	_, err = w.Write([]byte(body))
	if err != nil {
		l.Error("failed to write body", err)
	}
}
