// Package handlers holds the HTTP transport layer: decode, delegate, encode.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Response is the single response envelope for this API.
type Response struct {
	OK      bool              `json:"ok"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func writeJSON(w http.ResponseWriter, log *slog.Logger, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line is already sent, so this can only be logged.
		log.Error("encode response", slog.String("error", err.Error()))
	}
}
