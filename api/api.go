// Package api provides the useful functions, structures and variables for the api package
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var (
	// ErrInvalidStatusCode is returned when the status code is different from 200
	ErrInvalidStatusCode = errors.New("status code not ok")
)

// Error is the structure for the error response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

func send(w http.ResponseWriter, statusCode int, body interface{}) {
	const jsonContentType = "application/json; charset=utf-8"

	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("Unable to encode body as JSON", "error", err)
	}
}

func sendErr(w http.ResponseWriter, statusCode int, err error) {
	var httpErr Error
	if !errors.As(err, &httpErr) {
		httpErr = Error{
			Code:    "unknown_error",
			Message: "An unexpected error happened",
		}
	}
	if statusCode >= 500 {
		slog.Error("Unable to process request", "error", err.Error(), "status_code", statusCode)
	}

	send(w, statusCode, httpErr)
}
