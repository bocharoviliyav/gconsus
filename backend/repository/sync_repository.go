package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gconsus/database"
	"gconsus/entity"

	"github.com/google/uuid"
)

// SyncRepository handles sync_history and configurations database operations.
type SyncRepository struct {
	db *database.DB
}

// NewSyncRepository creates a new SyncRepository.
func NewSyncRepository(db *database.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// ---------------------------------------------------------------------------
// sync_history operations
// ---------------------------------------------------------------------------

// CreateSyncRecord inserts a new sync history row with status "running".
func (r *SyncRepository) CreateSyncRecord(ctx context.Context, record *entity.SyncHistory) error {
	query := `
		INSERT INTO sync_history (id, provider_id, sync_type, status, started_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	if record.StartedAt.IsZero() {
		record.StartedAt = time.Now()
	}
	if record.Status == "" {
		record.Status = entity.SyncStatusRunning
	}

	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.ProviderID, record.SyncType, record.Status, record.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}
	return nil
}

// CompleteSyncRecord marks a sync record as completed.
func (r *SyncRepository) CompleteSyncRecord(ctx context.Context, id uuid.UUID, usersSynced, activitiesSynced int) error {
	query := `
		UPDATE sync_history
		SET status = $1, users_synced = $2, activities_synced = $3, completed_at = $4
		WHERE id = $5
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, entity.SyncStatusCompleted, usersSynced, activitiesSynced, now, id)
	if err != nil {
		return fmt.Errorf("failed to complete sync record: %w", err)
	}
	return nil
}

// FailSyncRecord marks a sync record as failed with an error message.
func (r *SyncRepository) FailSyncRecord(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE sync_history
		SET status = $1, completed_at = $2, error_message = $3
		WHERE id = $4
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, entity.SyncStatusFailed, now, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark sync as failed: %w", err)
	}
	return nil
}

// GetLatestSync returns the most recent sync record of a given type.
func (r *SyncRepository) GetLatestSync(ctx context.Context, syncType entity.SyncType) (*entity.SyncHistory, error) {
	var record entity.SyncHistory
	query := `
		SELECT * FROM sync_history
		WHERE sync_type = $1
		ORDER BY started_at DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &record, query, syncType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sync: %w", err)
	}
	return &record, nil
}

// ListSyncHistory returns recent sync records.
func (r *SyncRepository) ListSyncHistory(ctx context.Context, limit int) ([]entity.SyncHistory, error) {
	query := `
		SELECT * FROM sync_history
		ORDER BY started_at DESC
		LIMIT $1
	`

	var records []entity.SyncHistory
	err := r.db.SelectContext(ctx, &records, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync history: %w", err)
	}
	return records, nil
}

// IsRunning checks if a sync of the given type is currently running.
func (r *SyncRepository) IsRunning(ctx context.Context, syncType entity.SyncType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sync_history
			WHERE sync_type = $1 AND status = $2
		)
	`

	var running bool
	err := r.db.GetContext(ctx, &running, query, syncType, entity.SyncStatusRunning)
	if err != nil {
		return false, fmt.Errorf("failed to check running sync: %w", err)
	}
	return running, nil
}

// DeleteOlderThan removes sync records older than the given date.
func (r *SyncRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	query := `DELETE FROM sync_history WHERE started_at < $1`

	result, err := r.db.ExecContext(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old sync records: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return rows, nil
}

// ---------------------------------------------------------------------------
// configurations operations
// ---------------------------------------------------------------------------

// GetConfig retrieves a configuration value by key.
func (r *SyncRepository) GetConfig(ctx context.Context, key string) (*entity.Configuration, error) {
	var cfg entity.Configuration
	query := `SELECT * FROM configurations WHERE key = $1`

	err := r.db.GetContext(ctx, &cfg, query, key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return &cfg, nil
}

// SetConfig inserts or updates a configuration value.
func (r *SyncRepository) SetConfig(ctx context.Context, key string, value entity.JSONValue, updatedBy *uuid.UUID) error {
	query := `
		INSERT INTO configurations (id, key, value, updated_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_by = EXCLUDED.updated_by
	`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), key, value, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}

// ListConfigs retrieves all configuration entries.
func (r *SyncRepository) ListConfigs(ctx context.Context) ([]entity.Configuration, error) {
	query := `SELECT * FROM configurations ORDER BY key`

	var configs []entity.Configuration
	err := r.db.SelectContext(ctx, &configs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	return configs, nil
}
