package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourorg/b25/services/media/internal/config"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS media (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			org_id VARCHAR(36),
			file_name VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			media_type VARCHAR(20) NOT NULL,
			size BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			storage_path TEXT NOT NULL,
			public_url TEXT NOT NULL,
			cdn_url TEXT,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_org_id (org_id),
			INDEX idx_media_type (media_type),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at)
		)`,

		`CREATE TABLE IF NOT EXISTS media_quota (
			id VARCHAR(36) PRIMARY KEY,
			entity_id VARCHAR(36) NOT NULL,
			type VARCHAR(20) NOT NULL,
			used BIGINT NOT NULL DEFAULT 0,
			"limit" BIGINT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(entity_id, type),
			INDEX idx_entity_id (entity_id)
		)`,

		`CREATE TABLE IF NOT EXISTS upload_sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			total_size BIGINT NOT NULL,
			chunk_size INT NOT NULL,
			total_chunks INT NOT NULL,
			uploaded_keys JSONB,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP NOT NULL,
			completed_at TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_status (status),
			INDEX idx_expires_at (expires_at)
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
