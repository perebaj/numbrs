package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNumbers(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"numbers":[1,2,3]}`))
		if err != nil {
			t.Fatal(err)
		}
	}))

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"numbers":[4,5,6]}`))
		if err != nil {
			t.Fatal(err)
		}
	}))

	target := "/numbers?u=" + server1.URL + "&u=" + server2.URL + "&u=" + "http://fake:8080"

	req := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	Handler(r)
	r.ServeHTTP(w, req)
	assert(t, http.StatusOK, w.Result().StatusCode)

	var gotResp numberResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&gotResp); err != nil {
		t.Fatal(err)
	}

	want := numberResponse{
		Numbers: []int{1, 2, 3, 4, 5, 6},
	}
	if len(gotResp.Numbers) == 6 {
		assert(t, want.Numbers[0], gotResp.Numbers[0])
		assert(t, want.Numbers[1], gotResp.Numbers[1])
		assert(t, want.Numbers[2], gotResp.Numbers[2])
		assert(t, want.Numbers[3], gotResp.Numbers[3])
		assert(t, want.Numbers[4], gotResp.Numbers[4])
		assert(t, want.Numbers[5], gotResp.Numbers[5])
	} else {
		t.Errorf("number of ints expected is wrong got %v want %v", len(gotResp.Numbers), len(want.Numbers))
	}
}

func TestNumbers_ErrorHandler(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	target := "/numbers?u=" + server1.URL + "&u=" + server2.URL

	req := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	Handler(r)
	r.ServeHTTP(w, req)
	assert(t, http.StatusInternalServerError, w.Result().StatusCode)

	var gotResp numberResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&gotResp); err != nil {
		t.Fatal(err)
	}
	if len(gotResp.Numbers) != 0 {
		t.Errorf("number of ints expected is wrong got %v want 0", len(gotResp.Numbers))
	}
}

func TestReq(t *testing.T) {
	want := testServerResponse{
		Numbers: []int{1, 2, 3},
	}

	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(wantBytes))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()
	ints, errs := request([]string{server.URL, server.URL})
	if len(errs) != 0 {
		t.Errorf("number of errors expected is wrong got %v want %v", len(errs), 0)
	}

	if len(ints) == 6 {
		// compare want to ints
		assert(t, want.Numbers[0], ints[0])
		assert(t, want.Numbers[1], ints[1])
		assert(t, want.Numbers[2], ints[2])
		assert(t, want.Numbers[0], ints[3])
		assert(t, want.Numbers[1], ints[4])
		assert(t, want.Numbers[2], ints[5])
	} else {
		t.Errorf("number of ints expected is wrong got %v want %v", len(ints), 6)
	}
}

func TestReq_InvalidStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	_, errs := request([]string{server.URL, server.URL})
	if len(errs) == 2 {
		assert(t, ErrInvalidStatusCode, errs[0])
		assert(t, ErrInvalidStatusCode, errs[1])
	} else {
		t.Errorf("number of errors expected is wrong got %v want %v", len(errs), 2)
	}
}

func assert(t *testing.T, want interface{}, got interface{}) {
	if want != got {
		t.Errorf("expected %v; got %v", want, got)
	}
}
