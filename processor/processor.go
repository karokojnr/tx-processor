package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"tx-processor/config"
	"tx-processor/models"
	"tx-processor/repository"
)

type Processor struct {
	cfg            *config.Config
	logger         *slog.Logger
	repo           repository.Analytics
	analyticsCache sync.Map // Thread-safe map for real-time data
	userMu         sync.Map // Per-user locks to prevent races

}

func NewProcessor(cfg *config.Config, logger *slog.Logger, repo repository.Analytics) *Processor {
	return &Processor{
		cfg:            cfg,
		logger:         logger,
		repo:           repo,
		analyticsCache: sync.Map{},
		userMu:         sync.Map{},
	}
}

func (p *Processor) ProcessStream(ctx context.Context, lines <-chan string, batchSize int) error {
	var batch []models.Transaction

	for line := range lines {
		var transaction models.Transaction
		if err := json.Unmarshal([]byte(line), &transaction); err != nil {
			p.logger.Warn("Skipping invalid JSON", "error", err)
			continue
		}

		batch = append(batch, transaction)

		if len(batch) >= batchSize {
			if err := p.applyTransactions(ctx, batch); err != nil {
				return fmt.Errorf("processing batch: %w", err)
			}
			batch = batch[:0] // Reset without reallocating
		}
	}

	if len(batch) > 0 {
		return p.applyTransactions(ctx, batch)
	}

	return nil
}

func (p *Processor) applyTransactions(ctx context.Context, txs []models.Transaction) error {
	localUpdates := make(map[string]*models.UserAnalytics)

	for _, tx := range txs {
		userID := tx.UserID

		// Per-user locking for thread safety without global contention
		lockIface, _ := p.userMu.LoadOrStore(userID, &sync.Mutex{})
		lock := lockIface.(*sync.Mutex)

		lock.Lock()

		dataIface, _ := p.analyticsCache.LoadOrStore(userID, &models.UserAnalytics{UserID: userID})
		userData := dataIface.(*models.UserAnalytics)

		value := tx.Price * float64(tx.Quantity)
		userData.TotalOrders++
		userData.TotalSpent += value

		if existing, ok := localUpdates[userID]; ok {
			existing.TotalOrders++
			existing.TotalSpent += value
		} else {
			localUpdates[userID] = &models.UserAnalytics{
				UserID:      userID,
				TotalOrders: 1,
				TotalSpent:  value,
			}
		}

		lock.Unlock()
	}

	if err := p.repo.UpdateAnalytics(ctx, localUpdates); err != nil {
		return err
	}

	p.logger.Info("batch processed",
		"transactions", len(txs),
		"users_affected", len(localUpdates))

	return nil
}
