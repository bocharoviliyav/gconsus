package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gconsus/database"
	"gconsus/entity"

	"github.com/google/uuid"
)

// ProviderRepository handles VCS provider database operations.
type ProviderRepository struct {
	db *database.DB
}

// NewProviderRepository creates a new ProviderRepository.
func NewProviderRepository(db *database.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// Create inserts a new VCS provider.
func (r *ProviderRepository) Create(ctx context.Context, p *entity.VCSProvider) error {
	query := `
		INSERT INTO vcs_providers (id, name, type, base_url, auth_token, enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	err := r.db.QueryRowxContext(ctx, query,
		p.ID, p.Name, p.Type, p.BaseURL, p.AuthToken, p.Enabled,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	return nil
}

// GetByID retrieves a provider by ID.
func (r *ProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.VCSProvider, error) {
	var p entity.VCSProvider
	query := `SELECT * FROM vcs_providers WHERE id = $1`

	err := r.db.GetContext(ctx, &p, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	return &p, nil
}

// List retrieves all providers, optionally filtered by enabled.
func (r *ProviderRepository) List(ctx context.Context, enabledOnly bool) ([]entity.VCSProvider, error) {
	query := `SELECT * FROM vcs_providers`
	args := []interface{}{}

	if enabledOnly {
		query += ` WHERE enabled = $1`
		args = append(args, true)
	}
	query += ` ORDER BY name`

	var providers []entity.VCSProvider
	err := r.db.SelectContext(ctx, &providers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	return providers, nil
}

// Update updates a provider.
func (r *ProviderRepository) Update(ctx context.Context, p *entity.VCSProvider) error {
	query := `
		UPDATE vcs_providers
		SET name = $1, type = $2, base_url = $3, auth_token = $4, enabled = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		p.Name, p.Type, p.BaseURL, p.AuthToken, p.Enabled, p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("provider not found")
	}
	return nil
}

// Delete removes a provider.
func (r *ProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vcs_providers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("provider not found")
	}
	return nil
}

// GetByType retrieves all providers of a given type (github/gitlab).
func (r *ProviderRepository) GetByType(ctx context.Context, providerType string) ([]entity.VCSProvider, error) {
	query := `SELECT * FROM vcs_providers WHERE type = $1 AND enabled = true ORDER BY name`

	var providers []entity.VCSProvider
	err := r.db.SelectContext(ctx, &providers, query, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers by type: %w", err)
	}
	return providers, nil
}
