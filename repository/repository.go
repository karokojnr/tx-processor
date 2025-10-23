package repository

import (
	"context"
	"database/sql"
	"fmt"
	"tx-processor/models"

	"github.com/jmoiron/sqlx"
)

// AnalyticsRepo provides methods for interacting with user analytics data.
// It implements Analytics interface
type AnalyticsRepo struct {
	db *sqlx.DB
}

// NewAnalyticsRepo creates a new AnalyticsRepo instance.
func NewAnalyticsRepo(db *sqlx.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

// UpdateAnalytics applies aggregated transaction updates for multiple users atomically.
func (r *AnalyticsRepo) UpdateAnalytics(ctx context.Context, updates map[string]*models.UserAnalytics) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("rollback error: %v\n", err)
		}
	}()

	// This query handles both new and existing users atomically
	query := `
    INSERT INTO user_analytics (user_id, total_orders, total_spent)
    VALUES ($1, $2, $3)
    ON CONFLICT(user_id) DO UPDATE SET
        total_orders = user_analytics.total_orders + EXCLUDED.total_orders,
        total_spent = user_analytics.total_spent + EXCLUDED.total_spent
    `

	stmt, err := tx.Preparex(query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, analytics := range updates {
		// Check if context was cancelled
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		if _, err := stmt.ExecContext(ctx, analytics.UserID, analytics.TotalOrders, analytics.TotalSpent); err != nil {
			return fmt.Errorf("exec update for user %s: %w", analytics.UserID, err)
		}
	}

	return tx.Commit()
}

// UserAnalytics retrieves analytics for a specific user
func (r *AnalyticsRepo) UserAnalytics(ctx context.Context, userID string) (*models.UserAnalytics, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	var analytics models.UserAnalytics
	query := "SELECT user_id, total_orders, total_spent FROM user_analytics WHERE user_id = $1"

	if err := r.db.GetContext(ctx, &analytics, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return &models.UserAnalytics{UserID: userID, TotalOrders: 0, TotalSpent: 0}, nil
		}
		return nil, fmt.Errorf("select user analytics: %w", err)
	}

	return &analytics, nil
}

// TopUsers returns top users ordered by total orders.
func (r *AnalyticsRepo) TopUsers(ctx context.Context, limit int) ([]models.UserAnalytics, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive, got %d", limit)
	}
	var users []models.UserAnalytics
	query := `
    SELECT user_id, total_orders, total_spent 
    FROM user_analytics 
    ORDER BY total_orders DESC 
    LIMIT $1
    `

	if err := r.db.SelectContext(ctx, &users, query, limit); err != nil {
		return nil, fmt.Errorf("select top users: %w", err)
	}

	return users, nil
}

// UserAnomalies returns users with anomalous activity based on order/spend deviation.
func (r *AnalyticsRepo) UserAnomalies(ctx context.Context) ([]models.AnomalyUser, error) {
	query := `
    WITH stats AS (
        SELECT 
            AVG(total_orders)::FLOAT as avg_orders,
            STDDEV(total_orders)::FLOAT as stddev_orders,
            AVG(total_spent)::FLOAT as avg_spent,
            STDDEV(total_spent)::FLOAT as stddev_spent
        FROM user_analytics
        WHERE total_orders > 0
    )
    SELECT 
        ua.user_id,
        ua.total_orders,
        ua.total_spent,
        (ua.total_orders > stats.avg_orders + 2 * stats.stddev_orders) as order_anomaly,
        (ua.total_spent > stats.avg_spent + 2 * stats.stddev_spent) as spending_anomaly
    FROM user_analytics ua, stats
    WHERE ua.total_orders > stats.avg_orders + 2 * stats.stddev_orders
       OR ua.total_spent > stats.avg_spent + 2 * stats.stddev_spent
    ORDER BY ua.total_orders DESC, ua.total_spent DESC
    `

	var anomalies []models.AnomalyUser
	if err := r.db.SelectContext(ctx, &anomalies, query); err != nil {
		return nil, fmt.Errorf("select anomalies: %w", err)
	}
	return anomalies, nil
}
