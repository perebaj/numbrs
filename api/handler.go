package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

// NumberResponse is the response from the /numbers endpoint
type NumberResponse struct {
	Numbers []int `json:"numbers"`
}

// TestServerResponse is the response from the test server
type TestServerResponse struct {
	Numbers []int    `json:"numbers"`
	Strings []string `json:"strings"`
}

// Handler gather all the handler functions
func Handler(r chi.Router) {
	r.HandleFunc("/numbers", numbers)
}

// numbers handles the /numbers endpoint
func numbers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	u := query["u"]

	if len(u) == 0 {
		sendErr(w, http.StatusBadRequest, Error{Code: "invalid_request", Message: "no valid urls available"})
	}

	intSlice, errs := request(u)
	// We want to return an error if all the urls are invalid
	if len(errs) == len(u) {
		err := errors.Join(errs...)
		slog.Error("all urls are invalid", "error", err)
		send(w, http.StatusInternalServerError, NumberResponse{
			Numbers: []int{},
		})
		return
	}

	intSlice = sortCompact(intSlice)

	send(w, http.StatusOK, NumberResponse{
		Numbers: intSlice,
	})
}

func sortCompact(intSlice []int) []int {
	slices.Sort(intSlice)
	intSlice = slices.Compact(intSlice)

	return intSlice
}

// request makes a request to the urls and returns the ints and errors
func request(u []string) ([]int, []error) {
	var errs []error
	var intResp []int
	for _, v := range u {
		client := http.Client{
			Timeout: 500 * time.Millisecond,
		}
		resp, err := client.Get(v)
		if err != nil {
			slog.Debug("bypassing http get error", "error", err)
			errs = append(errs, err)
			continue
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			errs = append(errs, ErrInvalidStatusCode)
		} else {
			var r TestServerResponse
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				errs = append(errs, err)
				continue
			}
			if len(r.Numbers) == 0 || len(r.Strings) != 0 {
				errs = append(errs, ErrInvalidResponse)
				continue
			}

			intResp = append(intResp, r.Numbers...)
		}
	}
	return intResp, errs
}
