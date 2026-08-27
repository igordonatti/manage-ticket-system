package httpapi

// import das bibliotecas?
import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type RedisChecker interface {
	Ping(ctx context.Context) error
}

type HealthResponse struct {
	Status  string `json:"status"`
	Redis   string `json:"redis"`
	Version string `json:"version"`
}

func NewHealthHandler(checker RedisChecker, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		statusCode := http.StatusOK

		response := HealthResponse{
			Status:  "ok",
			Redis:   "ok",
			Version: "dev",
		}

		if err := checker.Ping(ctx); err != nil {
			statusCode = http.StatusServiceUnavailable
			response.Status = "degraded"
			response.Redis = "unavailable"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		_ = json.NewEncoder(w).Encode(response)
	}
}
