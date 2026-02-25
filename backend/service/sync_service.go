package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gconsus/adapter/vcs"
	"gconsus/entity"
	"gconsus/repository"

	"github.com/google/uuid"
)

// SyncService orchestrates data synchronization from VCS providers.
type SyncService struct {
	activityRepo *repository.ActivityRepository
	syncRepo     *repository.SyncRepository
	userRepo     *repository.UserRepository
	settingsSvc  *SettingsService
	aggService   *AggregationService
}

// NewSyncService creates a new SyncService.
func NewSyncService(
	activityRepo *repository.ActivityRepository,
	syncRepo *repository.SyncRepository,
	userRepo *repository.UserRepository,
	settingsSvc *SettingsService,
	aggService *AggregationService,
) *SyncService {
	return &SyncService{
		activityRepo: activityRepo,
		syncRepo:     syncRepo,
		userRepo:     userRepo,
		settingsSvc:  settingsSvc,
		aggService:   aggService,
	}
}

// SyncStatus returns a flat status object expected by the frontend.
func (s *SyncService) SyncStatus(ctx context.Context) (map[string]interface{}, error) {
	gitSync, _ := s.syncRepo.GetLatestSync(ctx, entity.SyncTypeGitActivities)

	running, _ := s.syncRepo.IsRunning(ctx, entity.SyncTypeGitActivities)

	result := map[string]interface{}{
		"is_running":        running,
		"users_synced":      0,
		"teams_synced":      0,
		"activities_synced": 0,
	}

	if gitSync != nil {
		result["users_synced"] = gitSync.UsersSynced
		result["activities_synced"] = gitSync.ActivitiesSynced
		result["started_at"] = gitSync.StartedAt
		if gitSync.CompletedAt != nil {
			result["last_completed_at"] = gitSync.CompletedAt
		}
		if gitSync.ErrorMessage != nil {
			result["last_error"] = *gitSync.ErrorMessage
		}
	}

	return result, nil
}

// TriggerSync starts a full sync for all enabled providers.
// It is safe to call concurrently -- duplicate runs are prevented by a running check.
func (s *SyncService) TriggerSync(ctx context.Context) (*entity.SyncHistory, error) {
	running, err := s.syncRepo.IsRunning(ctx, entity.SyncTypeGitActivities)
	if err != nil {
		return nil, err
	}
	if running {
		return nil, fmt.Errorf("sync is already running")
	}

	// Create record
	record := &entity.SyncHistory{
		SyncType: entity.SyncTypeGitActivities,
	}
	if err := s.syncRepo.CreateSyncRecord(ctx, record); err != nil {
		return nil, err
	}

	// Run in background
	go func() {
		bgCtx := context.Background()
		usersSynced, activitiesSynced, syncErr := s.runFullSync(bgCtx)
		if syncErr != nil {
			slog.Error("sync: full sync failed", "error", syncErr)
			_ = s.syncRepo.FailSyncRecord(bgCtx, record.ID, syncErr.Error())
			return
		}
		_ = s.syncRepo.CompleteSyncRecord(bgCtx, record.ID, usersSynced, activitiesSynced)
		slog.Info("sync: full sync completed", "users", usersSynced, "activities", activitiesSynced)

		// Trigger aggregation so leaderboards and dashboard totals are available immediately.
		if activitiesSynced > 0 {
			slog.Info("sync: starting post-sync aggregation")
			for _, days := range []int{7, 30, 90, 180, 365} {
				if err := s.aggService.AggregateForLastNDays(bgCtx, days); err != nil {
					slog.Error("sync: aggregation failed", "days", days, "error", err)
				}
			}
			slog.Info("sync: post-sync aggregation completed")
		}
	}()

	return record, nil
}

// SyncHistory returns recent sync history entries.
func (s *SyncService) SyncHistory(ctx context.Context, limit int) ([]entity.SyncHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.syncRepo.ListSyncHistory(ctx, limit)
}

// ---------------------------------------------------------------------------
// Internal sync logic
// ---------------------------------------------------------------------------

func (s *SyncService) runFullSync(ctx context.Context) (int, int, error) {
	clients, err := s.settingsSvc.GetEnabledClients(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("get clients: %w", err)
	}

	if len(clients) == 0 {
		return 0, 0, fmt.Errorf("no enabled VCS providers configured")
	}

	// Step 1: Discover and upsert users from every enabled provider.
	for _, client := range clients {
		vcsUsers, err := client.FetchUsers(ctx)
		if err != nil {
			slog.Warn("sync: fetch users from provider failed", "error", err)
			continue
		}
		for _, vu := range vcsUsers {
			if vu.Username == "" {
				continue
			}
			parts := splitName(vu.Name)
			u := &entity.User{
			Username:  vu.Username,
				FirstName: parts[0],
				LastName:  parts[1],
				Email:     strPtr(vu.Email),
				PhotoURL:  strPtr(vu.AvatarURL),
				IsActive:  true,
			}
			if err := s.userRepo.Upsert(ctx, u); err != nil {
				slog.Warn("sync: upsert user", "login", vu.Username, "error", err)
			}
		}
		slog.Info("sync: discovered users from provider", "count", len(vcsUsers))
	}

	// Step 2: Sync activities for all active users.
	// Fetch 365 days so week / month / quarter / half-year / year views all have data.
	from := time.Now().AddDate(0, 0, -365)
	to := time.Now()

	activeFlag := true
	users, err := s.userRepo.List(ctx, &activeFlag, 0, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("list users: %w", err)
	}

	totalUsers := 0
	totalActivities := 0

	for providerID, client := range clients {
		for _, user := range users {
			login := user.Username
			if login == "" {
				continue
			}

			activities, err := client.FetchUserActivities(ctx, login, from, to)
			if err != nil {
				slog.Warn("sync: fetch activities failed",
					"provider", providerID, "user", login, "error", err)
				continue
			}

			if len(activities) == 0 {
				continue
			}

			dbActivities := make([]entity.GitActivity, 0, len(activities))
			for _, a := range activities {
				dbActivities = append(dbActivities, vcsActivityToEntity(a, user.ID, providerID))
			}

			if err := s.activityRepo.BatchCreate(ctx, dbActivities); err != nil {
				slog.Warn("sync: batch create failed",
					"provider", providerID, "user", login, "error", err)
				continue
			}

			totalActivities += len(dbActivities)
			totalUsers++
		}
	}

	return totalUsers, totalActivities, nil
}

// splitName splits "FirstName LastName" into [first, last]. Handles single names.
func splitName(full string) [2]string {
	parts := strings.SplitN(strings.TrimSpace(full), " ", 2)
	if len(parts) == 2 {
		return [2]string{parts[0], parts[1]}
	}
	if len(parts) == 1 && parts[0] != "" {
		return [2]string{parts[0], parts[0]}
	}
	return [2]string{"Unknown", "Unknown"}
}

// vcsActivityToEntity converts a VCS adapter activity to a database entity.
func vcsActivityToEntity(a vcs.Activity, userID, providerID uuid.UUID) entity.GitActivity {
	ga := entity.GitActivity{
		UserID:          userID,
		ProviderID:      providerID,
		ActivityType:    entity.ActivityType(a.Type),
		RepositoryName:  a.RepositoryName,
		RepositoryOwner: a.RepositoryOwner,
		CommitCount:     a.CommitCount,
		LinesAdded:      a.LinesAdded,
		LinesDeleted:    a.LinesDeleted,
		OccurredAt:      a.OccurredAt,
	}

	switch a.Type {
	case vcs.ActivityPR:
		ga.PRTitle = strPtr(a.Title)
		ga.PRURL = strPtr(a.URL)
		ga.PRMerged = a.Merged
	case vcs.ActivityIssue:
		ga.IssueTitle = strPtr(a.Title)
		ga.IssueURL = strPtr(a.URL)
		ga.IssueState = strPtr(a.State)
	case vcs.ActivityReview:
		ga.PRTitle = strPtr(a.Title)
		ga.PRURL = strPtr(a.URL)
	}

	return ga
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
