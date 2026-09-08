package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeRedisChecker struct {
	err error
}

func (f fakeRedisChecker) Ping(context.Context) error {
	return f.err
}

type blockingRedisChecker struct{}

func (blockingRedisChecker) Ping(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		checker    RedisChecker
		wantStatus int
		wantBody   HealthResponse
	}{
		{
			name:       "redis available",
			checker:    fakeRedisChecker{},
			wantStatus: http.StatusOK,
			wantBody: HealthResponse{
				Status:  "ok",
				Redis:   "ok",
				Version: "dev",
			},
		},
		{
			name:       "redis unavailable",
			checker:    fakeRedisChecker{err: errors.New("redis unavailable")},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: HealthResponse{
				Status:  "degraded",
				Redis:   "unavailable",
				Version: "dev",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(tt.checker, 100*time.Millisecond)
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status expected %d, received %d", tt.wantStatus, rr.Code)
			}

			if rr.Header().Get("Content-Type") != "application/json" {
				t.Fatalf(
					"Content-Type expected application/json, received %s",
					rr.Header().Get("Content-Type"),
				)
			}

			var got HealthResponse
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("error decoding JSON: %v", err)
			}

			if got != tt.wantBody {
				t.Errorf("expected response %+v, received %+v", tt.wantBody, got)
			}
		})
	}
}

func TestHealthHandlerTimesOut(t *testing.T) {
	handler := NewHealthHandler(blockingRedisChecker{}, time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"status expected %d, received %d",
			http.StatusServiceUnavailable,
			rr.Code,
		)
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler(fakeRedisChecker{}, 100*time.Millisecond)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status expected %d, received %d",
			http.StatusMethodNotAllowed,
			rr.Code,
		)
	}
}
