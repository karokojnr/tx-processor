package models

import "time"

// Transaction represents a single e-commerce order
type Transaction struct {
    OrderID   string    `json:"order_id"`
    UserID    string    `json:"user_id"`     // Key for aggregation
    ProductID string    `json:"product_id"`
    Quantity  int       `json:"quantity"`    // Affects total spending
    Price     float64   `json:"price"`       // Per-unit price
    Timestamp time.Time `json:"timestamp"`   // For time-based analysis
}

// UserAnalytics holds our real-time aggregated data
type UserAnalytics struct {
    UserID      string  `json:"user_id"`
    TotalOrders int     `json:"total_orders"`  // Count of transactions
    TotalSpent  float64 `json:"total_spent"`   // Sum of (price * quantity)
}

// Response is a generic API response wrapper
type Response[T any] struct {
    Data    *T     `json:"data,omitempty"`
    Message string `json:"message,omitempty"`
}