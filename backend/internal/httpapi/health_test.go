package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("statys expected %d, received %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf(
			"Content-Type expected to application/json, received %s",
			rr.Header().Get("Content-Type"))
	}

	var got HealthResponse

	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("error at decode JSON: %v", err)
	}

	want := HealthResponse{
		Status:  "ok",
		Redis:   "not_checked",
		Version: "dev",
	}

	if got != want {
		t.Errorf("expected response: %+v, received: %+v", want, got)
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status expect %d, received %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
