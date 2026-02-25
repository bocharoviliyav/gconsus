package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gconsus/entity"
	"gconsus/repository"

	"github.com/google/uuid"
)

// AggregationService handles metrics aggregation
type AggregationService struct {
	activityRepo *repository.ActivityRepository
	metricsRepo  *repository.MetricsRepository
	userRepo     *repository.UserRepository
	teamRepo     *repository.TeamRepository
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(
	activityRepo *repository.ActivityRepository,
	metricsRepo *repository.MetricsRepository,
	userRepo *repository.UserRepository,
	teamRepo *repository.TeamRepository,
) *AggregationService {
	return &AggregationService{
		activityRepo: activityRepo,
		metricsRepo:  metricsRepo,
		userRepo:     userRepo,
		teamRepo:     teamRepo,
	}
}

// AggregateUserMetrics calculates and stores metrics for a user
func (s *AggregationService) AggregateUserMetrics(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) error {
	slog.Info("Aggregating user metrics", "user_id", userID, "start", startDate, "end", endDate)

	// Get activities for the user
	activities, err := s.activityRepo.GetByUser(ctx, userID, startDate, endDate, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to get user activities: %w", err)
	}

	// Calculate metrics
	metrics := s.calculateMetrics(activities)
	metrics.UserID = &userID
	metrics.PeriodStart = startDate
	metrics.PeriodEnd = endDate

	// Calculate top repositories
	topRepos := s.calculateTopRepositories(activities, 10)
	if len(topRepos) > 0 {
		reposJSON := entity.JSONB{}
		for _, repo := range topRepos {
			key := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
			reposJSON[key] = map[string]interface{}{
				"commits":       repo.Commits,
				"lines_added":   repo.LinesAdded,
				"lines_deleted": repo.LinesDeleted,
				"prs":           repo.PullRequests,
				"reviews":       repo.Reviews,
			}
		}
		metrics.TopRepositories = reposJSON
	}

	// Save or update metrics
	if err := s.metricsRepo.Create(ctx, &metrics); err != nil {
		return fmt.Errorf("failed to save user metrics: %w", err)
	}

	slog.Info("User metrics aggregated successfully",
		"user_id", userID,
		"commits", metrics.TotalCommits,
		"lines_added", metrics.TotalLinesAdded,
		"repos", metrics.RepositoriesCount)

	return nil
}

// AggregateTeamMetrics calculates and stores metrics for a team
func (s *AggregationService) AggregateTeamMetrics(ctx context.Context, teamID uuid.UUID, startDate, endDate time.Time) error {
	slog.Info("Aggregating team metrics", "team_id", teamID, "start", startDate, "end", endDate)

	// Get activities for all team members
	activities, err := s.activityRepo.GetByTeam(ctx, teamID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get team activities: %w", err)
	}

	// Calculate metrics
	metrics := s.calculateMetrics(activities)
	metrics.TeamID = &teamID
	metrics.PeriodStart = startDate
	metrics.PeriodEnd = endDate

	// Calculate top repositories
	topRepos := s.calculateTopRepositories(activities, 10)
	if len(topRepos) > 0 {
		reposJSON := entity.JSONB{}
		for _, repo := range topRepos {
			key := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
			reposJSON[key] = map[string]interface{}{
				"commits":       repo.Commits,
				"lines_added":   repo.LinesAdded,
				"lines_deleted": repo.LinesDeleted,
				"prs":           repo.PullRequests,
				"reviews":       repo.Reviews,
			}
		}
		metrics.TopRepositories = reposJSON
	}

	// Save or update metrics
	if err := s.metricsRepo.Create(ctx, &metrics); err != nil {
		return fmt.Errorf("failed to save team metrics: %w", err)
	}

	slog.Info("Team metrics aggregated successfully",
		"team_id", teamID,
		"commits", metrics.TotalCommits,
		"members_activities", len(activities))

	return nil
}

// AggregateAllUsersForPeriod aggregates metrics for all active users
func (s *AggregationService) AggregateAllUsersForPeriod(ctx context.Context, startDate, endDate time.Time) error {
	slog.Info("Starting aggregation for all users", "start", startDate, "end", endDate)

	// Get all active users
	activeFlag := true
	users, err := s.userRepo.List(ctx, &activeFlag, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, user := range users {
		if err := s.AggregateUserMetrics(ctx, user.ID, startDate, endDate); err != nil {
			slog.Error("Failed to aggregate user metrics",
				"user_id", user.ID,
				"username", user.Username,
				"error", err)
			errorCount++
			continue
		}
		successCount++
	}

	slog.Info("Completed user aggregation",
		"total_users", len(users),
		"success", successCount,
		"errors", errorCount)

	return nil
}

// AggregateAllTeamsForPeriod aggregates metrics for all active teams
func (s *AggregationService) AggregateAllTeamsForPeriod(ctx context.Context, startDate, endDate time.Time) error {
	slog.Info("Starting aggregation for all teams", "start", startDate, "end", endDate)

	// Get all active teams
	activeFlag := true
	teams, err := s.teamRepo.List(ctx, &activeFlag, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get active teams: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, team := range teams {
		if err := s.AggregateTeamMetrics(ctx, team.ID, startDate, endDate); err != nil {
			slog.Error("Failed to aggregate team metrics",
				"team_id", team.ID,
				"name", team.Name,
				"error", err)
			errorCount++
			continue
		}
		successCount++
	}

	slog.Info("Completed team aggregation",
		"total_teams", len(teams),
		"success", successCount,
		"errors", errorCount)

	return nil
}

// AggregateForCurrentMonth aggregates metrics for the current month
func (s *AggregationService) AggregateForCurrentMonth(ctx context.Context) error {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	slog.Info("Aggregating for current month", "start", startDate, "end", endDate)

	if err := s.AggregateAllUsersForPeriod(ctx, startDate, endDate); err != nil {
		return fmt.Errorf("failed to aggregate users: %w", err)
	}

	if err := s.AggregateAllTeamsForPeriod(ctx, startDate, endDate); err != nil {
		return fmt.Errorf("failed to aggregate teams: %w", err)
	}

	return nil
}

// AggregateForLastNDays aggregates metrics for the last N days
func (s *AggregationService) AggregateForLastNDays(ctx context.Context, days int) error {
	now := time.Now().UTC()
	// Truncate to day boundaries so queries with the same period can find the rows.
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days)

	slog.Info("Aggregating for last N days", "days", days, "start", startDate, "end", endDate)

	if err := s.AggregateAllUsersForPeriod(ctx, startDate, endDate); err != nil {
		return fmt.Errorf("failed to aggregate users: %w", err)
	}

	if err := s.AggregateAllTeamsForPeriod(ctx, startDate, endDate); err != nil {
		return fmt.Errorf("failed to aggregate teams: %w", err)
	}

	return nil
}

// calculateMetrics calculates aggregated metrics from activities
func (s *AggregationService) calculateMetrics(activities []entity.GitActivity) entity.AggregatedMetrics {
	metrics := entity.AggregatedMetrics{}
	repoSet := make(map[string]bool)

	for _, activity := range activities {
		// Track unique repositories
		repoKey := fmt.Sprintf("%s/%s", activity.RepositoryOwner, activity.RepositoryName)
		repoSet[repoKey] = true

		// Aggregate by activity type
		switch activity.ActivityType {
		case entity.ActivityTypeCommit:
			metrics.TotalCommits += activity.CommitCount
			metrics.TotalLinesAdded += int64(activity.LinesAdded)
			metrics.TotalLinesDeleted += int64(activity.LinesDeleted)

		case entity.ActivityTypePR:
			metrics.TotalPRs++
			if activity.PRMerged != nil && *activity.PRMerged {
				metrics.TotalPRsMerged++
			}

		case entity.ActivityTypeReview:
			metrics.TotalReviews++

		case entity.ActivityTypeIssue:
			metrics.TotalIssues++
		}
	}

	metrics.RepositoriesCount = len(repoSet)

	return metrics
}

// calculateTopRepositories calculates top N repositories by activity
func (s *AggregationService) calculateTopRepositories(activities []entity.GitActivity, topN int) []entity.RepositoryStats {
	repoStats := make(map[string]*entity.RepositoryStats)

	for _, activity := range activities {
		repoKey := fmt.Sprintf("%s/%s", activity.RepositoryOwner, activity.RepositoryName)

		if _, exists := repoStats[repoKey]; !exists {
			repoStats[repoKey] = &entity.RepositoryStats{
				Owner: activity.RepositoryOwner,
				Name:  activity.RepositoryName,
			}
		}

		repo := repoStats[repoKey]

		switch activity.ActivityType {
		case entity.ActivityTypeCommit:
			repo.Commits += activity.CommitCount
			repo.LinesAdded += int64(activity.LinesAdded)
			repo.LinesDeleted += int64(activity.LinesDeleted)

		case entity.ActivityTypePR:
			repo.PullRequests++

		case entity.ActivityTypeReview:
			repo.Reviews++
		}
	}

	// Convert map to slice
	var repoList []entity.RepositoryStats
	for _, stats := range repoStats {
		repoList = append(repoList, *stats)
	}

	// Sort by commits (simple bubble sort for top N)
	for i := 0; i < len(repoList); i++ {
		for j := i + 1; j < len(repoList); j++ {
			if repoList[j].Commits > repoList[i].Commits {
				repoList[i], repoList[j] = repoList[j], repoList[i]
			}
		}
	}

	// Return top N
	if len(repoList) > topN {
		return repoList[:topN]
	}
	return repoList
}

// CalculateScore calculates a performance score for a user
func (s *AggregationService) CalculateScore(metrics *entity.AggregatedMetrics) float64 {
	// Score formula:
	// commits * 1.0 + lines_added * 0.01 + prs_merged * 5.0 + reviews * 3.0 + issues * 2.0
	score := float64(metrics.TotalCommits)*1.0 +
		float64(metrics.TotalLinesAdded)*0.01 +
		float64(metrics.TotalPRsMerged)*5.0 +
		float64(metrics.TotalReviews)*3.0 +
		float64(metrics.TotalIssues)*2.0

	return score
}
