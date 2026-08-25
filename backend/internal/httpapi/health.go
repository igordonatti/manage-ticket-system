package httpapi

// import das bibliotecas?
import (
	"encoding/json"
	"net/http"
)

// objeto de resposta name/type/json opbject
type HealthResponse struct {
	Status  string `json:"status"`
	Redis   string `json:"redis"`
	Version string `json:"version"`
}

// handle endpoint /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := HealthResponse{
		Status:  "ok",
		Redis:   "not_checked",
		Version: "dev",
	}

	// escreve o struct montado em JSON na resposta HTTP
	_ = json.NewEncoder(w).Encode(resp)
}
