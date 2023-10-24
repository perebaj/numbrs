// Package main gather all the packages and start the server
package main

import (
	"net/http"
	"syscall"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/perebaj/numbrs/api"
)

func main() {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		api.Handler(r)
	})

	srv := http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("Server started on port 8080")

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error("server failed to start", "error", err)
		syscall.Exit(1)
	}
}
