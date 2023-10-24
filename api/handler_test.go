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
		_, err := w.Write([]byte(`{"numbers":[3,1,3,4]}`))
		if err != nil {
			t.Fatal(err)
		}
	}))

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"numbers":[4,2,6,4]}`))
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
		Numbers: []int{1, 2, 3, 4, 6},
	}
	if len(gotResp.Numbers) == 5 {
		assert(t, want.Numbers[0], gotResp.Numbers[0])
		assert(t, want.Numbers[1], gotResp.Numbers[1])
		assert(t, want.Numbers[2], gotResp.Numbers[2])
		assert(t, want.Numbers[3], gotResp.Numbers[3])
		assert(t, want.Numbers[4], gotResp.Numbers[4])
	} else {
		t.Errorf("number of ints expected is wrong got %v want %v", len(gotResp.Numbers), len(want.Numbers))
	}
}

func TestNumbers_ErrorHandler(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"wrong":[4,5,6]}`))
		if err != nil {
			t.Fatal(err)
		}
	}))

	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	target := "/numbers?u=" + server1.URL + "&u=" + server2.URL + "&u=" + server3.URL

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

func TestNumbers_NoValidURLs(t *testing.T) {
	target := "/numbers"

	req := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	Handler(r)
	r.ServeHTTP(w, req)
	assert(t, http.StatusBadRequest, w.Result().StatusCode)

	var gotResp Error
	if err := json.NewDecoder(w.Result().Body).Decode(&gotResp); err != nil {
		t.Fatal(err)
	}
	assert(t, "invalid_request", gotResp.Code)
	assert(t, "no valid urls available", gotResp.Message)
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
	intSlice, errs := request([]string{server.URL, server.URL})
	if len(errs) != 0 {
		t.Errorf("number of errors expected is wrong got %v want %v", len(errs), 0)
	}

	if len(intSlice) == 6 {
		// compare want to ints
		assert(t, want.Numbers[0], intSlice[0])
		assert(t, want.Numbers[1], intSlice[1])
		assert(t, want.Numbers[2], intSlice[2])
		assert(t, want.Numbers[0], intSlice[3])
		assert(t, want.Numbers[1], intSlice[4])
		assert(t, want.Numbers[2], intSlice[5])
	} else {
		t.Errorf("number of ints expected is wrong got %v want %v", len(intSlice), 6)
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

func TestSortCompact(t *testing.T) {
	want := []int{1, 2, 3, 4, 5, 6, 9, 10, 55}
	got := sortCompact([]int{9, 10, 55, 1, 2, 3, 4, 55, 5, 6, 1, 2, 3})

	if len(got) == 9 {
		assert(t, want[0], got[0])
		assert(t, want[1], got[1])
		assert(t, want[2], got[2])
		assert(t, want[3], got[3])
		assert(t, want[4], got[4])
		assert(t, want[5], got[5])
		assert(t, want[6], got[6])
		assert(t, want[7], got[7])
		assert(t, want[8], got[8])
	} else {
		t.Errorf("number of ints expected is wrong got %v want %v", len(got), 9)
	}
}

func assert(t *testing.T, want interface{}, got interface{}) {
	if want != got {
		t.Errorf("expected %v; got %v", want, got)
	}
}
