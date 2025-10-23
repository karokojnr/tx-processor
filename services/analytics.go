package services

import (
	"context"
	"fmt"
	"tx-processor/cache"
	"tx-processor/models"
)

// Analytics defines methods for user analytics operations
type Analytics interface {
	UpdateAnalytics(ctx context.Context, updates map[string]*models.UserAnalytics) error
	UserAnalytics(ctx context.Context, userID string) (*models.UserAnalytics, error)
	TopUsers(ctx context.Context, limit int) ([]models.UserAnalytics, error)
	UserAnomalies(ctx context.Context) ([]models.AnomalyUser, error)
}

type AnalyticsService struct {
	repo  Analytics
	cache cache.AnalyticsCache
}

func NewAnalyticsService(repo Analytics, cache cache.AnalyticsCache) *AnalyticsService {
	return &AnalyticsService{
		repo:  repo,
		cache: cache,
	}
}

// GetUserAnalytics retrieves user analytics with cache-first strategy
func (s *AnalyticsService) GetUserAnalytics(ctx context.Context, userID string) (*models.UserAnalytics, error) {
	// Try cache first
	if analytics, err := s.cache.Get(ctx, userID); err == nil {
		return analytics, nil
	}

	// Fallback to database
	analytics, err := s.repo.UserAnalytics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user analytics from repository: %w", err)
	}

	// Cache the result asynchronously (fire and forget)
	go func() {
		s.cache.Set(context.Background(), *analytics)
	}()

	return analytics, nil
}

// GetUserTotalOrders gets total orders for a specific user
func (s *AnalyticsService) GetUserTotalOrders(ctx context.Context, userID string) (int, error) {
	analytics, err := s.GetUserAnalytics(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user total orders: %w", err)
	}
	return analytics.TotalOrders, nil
}

// GetUserTotalSpendings gets total spendings for a specific user
func (s *AnalyticsService) GetUserTotalSpendings(ctx context.Context, userID string) (float64, error) {
	analytics, err := s.GetUserAnalytics(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user total spendings: %w", err)
	}
	return analytics.TotalSpent, nil
}

// GetTopUsers retrieves top users by orders
func (s *AnalyticsService) GetTopUsers(ctx context.Context, limit int) ([]models.UserAnalytics, error) {
	// Get from database (cache can be added later if needed)
	users, err := s.repo.TopUsers(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users from repository: %w", err)
	}

	return users, nil
}

// DetectAnomalies performs anomaly detection using the repository's implementation
func (s *AnalyticsService) DetectAnomalies(ctx context.Context) ([]models.AnomalyUser, error) {
	// Use the repository's anomaly detection logic
	anomalies, err := s.repo.UserAnomalies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect anomalies: %w", err)
	}

	return anomalies, nil
}

// InvalidateUserCache removes user data from cache (useful after updates)
func (s *AnalyticsService) InvalidateUserCache(ctx context.Context, userID string) error {
	if err := s.cache.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}
	return nil
}
