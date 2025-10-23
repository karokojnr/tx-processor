package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tx-processor/cache"
	"tx-processor/config"
	"tx-processor/repository"
)

type Handler struct {
	repo  repository.Analytics
	cache cache.AnalyticsCache
	cfg   *config.Config
}

func NewHandler(repo repository.Analytics, cache cache.AnalyticsCache, cfg *config.Config) *Handler {
	return &Handler{
		repo:  repo,
		cache: cache,
		cfg:   cfg,
	}
}

func (h *Handler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("/total_orders", h.totalOrdersHandler())
	r.HandleFunc("/total_spendings", h.totalSpendingsHandler())
	r.HandleFunc("/top_users", h.topUsersHandler())
	r.HandleFunc("/anomalies", h.anomaliesHandler())
}

func writeJSONResponse[T any](w http.ResponseWriter, status int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func writeErrorResponse(w http.ResponseWriter, status int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	return writeJSONResponse(w, status, errResp)

}

func readJSON[T any](r *http.Request) (T, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("decode json: %w", err)
	}
	return data, nil
}
