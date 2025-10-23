package cache

import (
	"context"
	"tx-processor/models"
)

type AnalyticsCache interface {
	Get(ctx context.Context, userID string) (*models.UserAnalytics, error)
	Set(ctx context.Context, analytics models.UserAnalytics) error
	Delete(ctx context.Context, userID string) error
}
