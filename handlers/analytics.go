package handlers

import (
	"fmt"
	"strconv"
	"tx-processor/repository"

	"net/http"
)

type AnalyticsHandler struct {
	repo repository.Analytics
}

func NewAnalyticsHandler(repo repository.Analytics) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("/total_orders", h.totalOrdersHandler())
	r.HandleFunc("/total_spendings", h.totalSpendingsHandler())
	r.HandleFunc("/top_users", h.topUsersHandler())
	r.HandleFunc("/anomalies", h.anomaliesHandler())
}

func (h *AnalyticsHandler) totalOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {

			writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required")
			return
		}
		analytics, err := h.repo.UserAnalytics(r.Context(), userID)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "failed to get user analytics")
			return
		}

		response := struct {
			UserID      string `json:"user_id"`
			TotalOrders int    `json:"total_orders"`
			Message     string `json:"message"`
		}{
			UserID:      userID,
			TotalOrders: analytics.TotalOrders,
			Message:     fmt.Sprintf("User %s has placed %d orders", userID, analytics.TotalOrders),
		}

		if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		}
	}
}

func (h *AnalyticsHandler) totalSpendingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required")
			return
		}
		analytics, err := h.repo.UserAnalytics(r.Context(), userID)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "failed to get user analytics")
			return
		}

		response := struct {
			UserID     string  `json:"user_id"`
			TotalSpent float64 `json:"total_spent"`
			Message    string  `json:"message"`
		}{
			UserID:     userID,
			TotalSpent: analytics.TotalSpent,
			Message:    fmt.Sprintf("User %s has spent a total of %.2f", userID, analytics.TotalSpent),
		}

		if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		}
	}
}

func (h *AnalyticsHandler) topUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitParam := r.URL.Query().Get("limit")
		limit := 10
		if limitParam != "" {
			l, err := strconv.Atoi(limitParam)
			if err != nil || l <= 0 {
				writeErrorResponse(w, http.StatusBadRequest, "limit must be a positive integer")
				return
			}
			limit = l
		}

		users, err := h.repo.TopUsers(r.Context(), limit)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "failed to get top users")
			return
		}

		if err := writeJSONResponse(w, http.StatusOK, users); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		}
	}
}

func (h *AnalyticsHandler) anomaliesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		anomalies, err := h.repo.UserAnomalies(r.Context())
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "failed to detect anomalies")
			return
		}

		if err := writeJSONResponse(w, http.StatusOK, anomalies); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
		}
	}
}
