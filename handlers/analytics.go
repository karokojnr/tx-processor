package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) totalOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required")
			return
		}

		analytics, err := h.analyticsService.GetUserAnalytics(ctx, userID)
		if err != nil {
			h.logger.Error("failed to get user analytics", "user_id", userID, "error", err)
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
			h.logger.Error("failed to write response", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to encode response")
		}
	}
}

func (h *Handler) totalSpendingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required")
			return
		}

		analytics, err := h.analyticsService.GetUserAnalytics(ctx, userID)
		if err != nil {
			h.logger.Error("failed to get user analytics", "user_id", userID, "error", err)
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
			Message:    fmt.Sprintf("User %s has spent $%.2f", userID, analytics.TotalSpent),
		}

		if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
			h.logger.Error("failed to write response", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to encode response")
		}
	}
}

func (h *Handler) topUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		limitStr := r.URL.Query().Get("limit")
		limit := 10 // default limit
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		users, err := h.analyticsService.GetTopUsers(ctx, limit)
		if err != nil {
			h.logger.Error("failed to get top users", "limit", limit, "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to get top users")
			return
		}

		response := struct {
			Users   []interface{} `json:"users"`
			Count   int           `json:"count"`
			Message string        `json:"message"`
		}{
			Users:   make([]interface{}, len(users)),
			Count:   len(users),
			Message: fmt.Sprintf("Retrieved top %d users by orders", len(users)),
		}

		for i, user := range users {
			response.Users[i] = struct {
				UserID      string  `json:"user_id"`
				TotalOrders int     `json:"total_orders"`
				TotalSpent  float64 `json:"total_spent"`
			}{
				UserID:      user.UserID,
				TotalOrders: user.TotalOrders,
				TotalSpent:  user.TotalSpent,
			}
		}

		if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
			h.logger.Error("failed to write response", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to encode response")
		}
	}
}

func (h *Handler) anomaliesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		anomalies, err := h.analyticsService.DetectAnomalies(ctx)
		if err != nil {
			h.logger.Error("failed to detect anomalies", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to detect anomalies")
			return
		}

		response := struct {
			Anomalies []interface{} `json:"anomalies"`
			Count     int           `json:"count"`
			Message   string        `json:"message"`
		}{
			Anomalies: make([]interface{}, len(anomalies)),
			Count:     len(anomalies),
			Message:   fmt.Sprintf("Detected %d anomalous users", len(anomalies)),
		}

		for i, anomaly := range anomalies {
			response.Anomalies[i] = struct {
				UserID          string  `json:"user_id"`
				TotalOrders     int     `json:"total_orders"`
				TotalSpent      float64 `json:"total_spent"`
				OrderAnomaly    bool    `json:"order_anomaly"`
				SpendingAnomaly bool    `json:"spending_anomaly"`
			}{
				UserID:          anomaly.UserID,
				TotalOrders:     anomaly.TotalOrders,
				TotalSpent:      anomaly.TotalSpent,
				OrderAnomaly:    anomaly.OrderAnomaly,
				SpendingAnomaly: anomaly.SpendingAnomaly,
			}
		}

		if err := writeJSONResponse(w, http.StatusOK, response); err != nil {
			h.logger.Error("failed to write response", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "failed to encode response")
		}
	}
}
