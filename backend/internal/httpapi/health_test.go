package httpapi

import (
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
		t.Fatalf("Content-Type expected to application/json, received %s", rr.Header().Get("Content-Type"))
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
