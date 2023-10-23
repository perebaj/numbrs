package api

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

type numberResponse struct {
	Numbers []int `json:"numbers"`
}

type testServerResponse struct {
	Numbers []int    `json:"numbers"`
	Strings []string `json:"strings"`
}

// Handler gather all the handler functions
func Handler(r chi.Router) {
	r.HandleFunc("/numbers", numbers)
}

func numbers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	u := query["u"]

	if len(u) == 0 {
		sendErr(w, http.StatusBadRequest, Error{Code: "invalid_request", Message: "no valid urls available"})
	}

	ints, errs := request(u)
	// We want to return an error if all the urls are invalid
	if len(errs) == len(u) {
		send(w, http.StatusInternalServerError, numberResponse{
			Numbers: []int{},
		})
	} else {
		send(w, http.StatusOK, numberResponse{
			Numbers: ints,
		})
	}
}

func request(u []string) ([]int, []error) {
	var errs []error
	var intResp []int
	for _, v := range u {
		resp, err := http.Get(v)
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
			var r testServerResponse
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				slog.Debug("bypassing status code different from 200", "error", err)
			}
			intResp = append(intResp, r.Numbers...)
		}
	}
	return intResp, errs
}
