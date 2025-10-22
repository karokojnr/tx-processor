package db

import (
	"fmt"
	"tx-processor/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Optimize connection pool for high throughput
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

func createSchema(db *sqlx.DB) error {
	schema := `
    -- Our main analytics table
    CREATE TABLE IF NOT EXISTS user_analytics (
        user_id VARCHAR(255) PRIMARY KEY,
        total_orders INTEGER DEFAULT 0,
        total_spent DECIMAL(15,2) DEFAULT 0.0,
        last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    -- These indexes are crucial for our API performance!
    CREATE INDEX IF NOT EXISTS idx_user_analytics_orders 
    ON user_analytics(total_orders DESC);
    
    CREATE INDEX IF NOT EXISTS idx_user_analytics_spent 
    ON user_analytics(total_spent DESC);
    
    -- Automatic timestamp updates
    CREATE OR REPLACE FUNCTION update_last_updated_column()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.last_updated = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;
    $$ language 'plpgsql';

    DROP TRIGGER IF EXISTS update_user_analytics_last_updated ON user_analytics;
    CREATE TRIGGER update_user_analytics_last_updated
        BEFORE UPDATE ON user_analytics
        FOR EACH ROW
        EXECUTE FUNCTION update_last_updated_column();
    `

	_, err := db.Exec(schema)
	return err
}
