package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"gconsus/adapter/github"
	glAdapter "gconsus/adapter/gitlab"
	"gconsus/adapter/vcs"
	"gconsus/entity"
	"gconsus/repository"

	"github.com/google/uuid"
)

// SettingsService manages VCS providers and system configuration.
type SettingsService struct {
	providerRepo *repository.ProviderRepository
	syncRepo     *repository.SyncRepository
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(
	providerRepo *repository.ProviderRepository,
	syncRepo *repository.SyncRepository,
) *SettingsService {
	return &SettingsService{
		providerRepo: providerRepo,
		syncRepo:     syncRepo,
	}
}

// ---------------------------------------------------------------------------
// VCS Provider management
// ---------------------------------------------------------------------------

// ListProviders returns all registered VCS providers.
func (s *SettingsService) ListProviders(ctx context.Context, enabledOnly bool) ([]entity.VCSProvider, error) {
	return s.providerRepo.List(ctx, enabledOnly)
}

// GetProvider returns a single provider by ID.
func (s *SettingsService) GetProvider(ctx context.Context, id uuid.UUID) (*entity.VCSProvider, error) {
	return s.providerRepo.GetByID(ctx, id)
}

// CreateProvider registers a new VCS provider.
func (s *SettingsService) CreateProvider(ctx context.Context, p *entity.VCSProvider) error {
	if p.Type != "github" && p.Type != "gitlab" {
		return fmt.Errorf("unsupported provider type: %s", p.Type)
	}
	return s.providerRepo.Create(ctx, p)
}

// UpdateProvider updates a provider.
func (s *SettingsService) UpdateProvider(ctx context.Context, p *entity.VCSProvider) error {
	return s.providerRepo.Update(ctx, p)
}

// DeleteProvider removes a provider.
func (s *SettingsService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	return s.providerRepo.Delete(ctx, id)
}

// TestProviderConnection tests the token against the VCS provider.
func (s *SettingsService) TestProviderConnection(ctx context.Context, id uuid.UUID) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	client, err := s.buildClient(provider)
	if err != nil {
		return err
	}
	return client.TestConnection(ctx)
}

// BuildClientForProvider creates a vcs.Client from a stored provider row.
func (s *SettingsService) BuildClientForProvider(ctx context.Context, id uuid.UUID) (vcs.Client, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return s.buildClient(provider)
}

// GetEnabledClients returns a vcs.Client for every enabled provider.
func (s *SettingsService) GetEnabledClients(ctx context.Context) (map[uuid.UUID]vcs.Client, error) {
	providers, err := s.providerRepo.List(ctx, true)
	if err != nil {
		return nil, err
	}

	clients := make(map[uuid.UUID]vcs.Client, len(providers))
	for _, p := range providers {
		c, err := s.buildClient(&p)
		if err != nil {
			slog.Warn("settings: cannot build client", "provider", p.Name, "error", err)
			continue
		}
		clients[p.ID] = c
	}
	return clients, nil
}

func (s *SettingsService) buildClient(p *entity.VCSProvider) (vcs.Client, error) {
	switch p.Type {
	case "github":
		cfg := github.DefaultConfig()
		cfg.BaseURL = p.BaseURL
		cfg.AuthToken = p.AuthToken
		c, err := github.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("github client: %w", err)
		}
		return &c, nil
	case "gitlab":
		cfg := glAdapter.DefaultConfig()
		cfg.BaseURL = p.BaseURL
		cfg.AuthToken = p.AuthToken
		c, err := glAdapter.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("gitlab client: %w", err)
		}
		return &c, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}

// ---------------------------------------------------------------------------
// Settings update from frontend
// ---------------------------------------------------------------------------

// UpsertProviderSettings creates or updates a VCS provider from the settings UI.
// If a provider of the given type already exists, it is updated; otherwise a new one is created.
// If authToken is empty the existing token is preserved (frontend sends empty when unchanged).
func (s *SettingsService) UpsertProviderSettings(ctx context.Context, providerType string, enabled bool, baseURL, authToken string) error {
	existing, _ := s.providerRepo.GetByType(ctx, providerType)
	if len(existing) > 0 {
		p := existing[0]
		p.Enabled = enabled
		p.BaseURL = baseURL
		if authToken != "" {
			p.AuthToken = authToken
		}
		return s.providerRepo.Update(ctx, &p)
	}
	p := &entity.VCSProvider{
		Name:      providerType,
		Type:      providerType,
		BaseURL:   baseURL,
		AuthToken: authToken,
		Enabled:   enabled,
	}
	return s.providerRepo.Create(ctx, p)
}

// TestConnectionDirect tests a VCS connection without persisting anything to the DB.
func (s *SettingsService) TestConnectionDirect(ctx context.Context, providerType, baseURL, authToken string) error {
	p := &entity.VCSProvider{
		Type:      providerType,
		BaseURL:   baseURL,
		AuthToken: authToken,
	}
	client, err := s.buildClient(p)
	if err != nil {
		return err
	}
	return client.TestConnection(ctx)
}

// ---------------------------------------------------------------------------
// System configuration
// ---------------------------------------------------------------------------

// GetConfig returns a single configuration by key.
func (s *SettingsService) GetConfig(ctx context.Context, key string) (*entity.Configuration, error) {
	return s.syncRepo.GetConfig(ctx, key)
}

// SetConfig sets a configuration value.
func (s *SettingsService) SetConfig(ctx context.Context, key string, value entity.JSONValue, updatedBy *uuid.UUID) error {
	return s.syncRepo.SetConfig(ctx, key, value, updatedBy)
}

// ListConfigs returns all configuration entries.
func (s *SettingsService) ListConfigs(ctx context.Context) ([]entity.Configuration, error) {
	return s.syncRepo.ListConfigs(ctx)
}

// GetSystemInfo returns a flat settings object expected by the frontend.
func (s *SettingsService) GetSystemInfo(ctx context.Context) (map[string]interface{}, error) {
	providers, err := s.providerRepo.List(ctx, false)
	if err != nil {
		return nil, err
	}

	configs, err := s.syncRepo.ListConfigs(ctx)
	if err != nil {
		return nil, err
	}

	latestSync, _ := s.syncRepo.GetLatestSync(ctx, entity.SyncTypeGitActivities)

	// Build flat structure expected by the frontend (SystemSettings).
	result := map[string]interface{}{
		"github_enabled":      false,
		"github_url":          "",
		"github_token_set":    false,
		"gitlab_enabled":      false,
		"gitlab_url":          "",
		"gitlab_token_set":    false,
		"sync_schedule_cron":  "0 */6 * * *",
		"sync_enabled":        true,
		"version":             "1.0.0",
	}

	// Populate provider fields.
	for _, p := range providers {
		switch p.Type {
		case "github":
			result["github_enabled"] = p.Enabled
			result["github_url"] = p.BaseURL
			result["github_token_set"] = p.AuthToken != ""
			result["updated_at"] = p.UpdatedAt
		case "gitlab":
			result["gitlab_enabled"] = p.Enabled
			result["gitlab_url"] = p.BaseURL
			result["gitlab_token_set"] = p.AuthToken != ""
			if _, ok := result["updated_at"]; !ok {
				result["updated_at"] = p.UpdatedAt
			}
		}
	}

	// Populate configuration fields.
	for _, cfg := range configs {
		switch cfg.Key {
		case "sync_schedule":
			// Value is stored as raw JSON (e.g. `"0 */6 * * *"`).
			var s string
			if err := json.Unmarshal(cfg.Value, &s); err == nil {
				result["sync_schedule_cron"] = s
			}
		case "sync_enabled":
			var b bool
			if err := json.Unmarshal(cfg.Value, &b); err == nil {
				result["sync_enabled"] = b
			}
		}
	}

	if latestSync != nil && latestSync.CompletedAt != nil {
		result["last_sync_at"] = latestSync.CompletedAt
	}

	return result, nil
}
