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

// MetricsRepository handles aggregated metrics database operations
type MetricsRepository struct {
	db *database.DB
}

// NewMetricsRepository creates a new MetricsRepository
func NewMetricsRepository(db *database.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// Create creates a new aggregated metric record
func (r *MetricsRepository) Create(ctx context.Context, metric *entity.AggregatedMetrics) error {
	query := `
		INSERT INTO aggregated_metrics (
			id, user_id, team_id, period_start, period_end,
			total_commits, total_lines_added, total_lines_deleted,
			total_prs, total_prs_merged, total_reviews, total_issues,
			repositories_count, top_repositories
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (user_id, period_start, period_end) WHERE user_id IS NOT NULL
		DO UPDATE SET
			total_commits = EXCLUDED.total_commits,
			total_lines_added = EXCLUDED.total_lines_added,
			total_lines_deleted = EXCLUDED.total_lines_deleted,
			total_prs = EXCLUDED.total_prs,
			total_prs_merged = EXCLUDED.total_prs_merged,
			total_reviews = EXCLUDED.total_reviews,
			total_issues = EXCLUDED.total_issues,
			repositories_count = EXCLUDED.repositories_count,
			top_repositories = EXCLUDED.top_repositories,
			created_at = NOW()
		RETURNING id, created_at
	`

	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}

	err := r.db.QueryRowxContext(ctx, query,
		metric.ID, metric.UserID, metric.TeamID, metric.PeriodStart, metric.PeriodEnd,
		metric.TotalCommits, metric.TotalLinesAdded, metric.TotalLinesDeleted,
		metric.TotalPRs, metric.TotalPRsMerged, metric.TotalReviews, metric.TotalIssues,
		metric.RepositoriesCount, metric.TopRepositories,
	).Scan(&metric.ID, &metric.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}

	return nil
}

// GetByUserAndPeriod retrieves metrics for a user in a specific period
func (r *MetricsRepository) GetByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) (*entity.AggregatedMetrics, error) {
	var metric entity.AggregatedMetrics
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id = $1 AND period_start = $2 AND period_end = $3
	`

	err := r.db.GetContext(ctx, &metric, query, userID, start, end)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metrics not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return &metric, nil
}

// GetByTeamAndPeriod retrieves metrics for a team in a specific period
func (r *MetricsRepository) GetByTeamAndPeriod(ctx context.Context, teamID uuid.UUID, start, end time.Time) (*entity.AggregatedMetrics, error) {
	var metric entity.AggregatedMetrics
	query := `
		SELECT * FROM aggregated_metrics
		WHERE team_id = $1 AND period_start = $2 AND period_end = $3
	`

	err := r.db.GetContext(ctx, &metric, query, teamID, start, end)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metrics not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return &metric, nil
}

// GetUserMetricsHistory retrieves historical metrics for a user
func (r *MetricsRepository) GetUserMetricsHistory(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id = $1
		ORDER BY period_start DESC
		LIMIT $2
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics history: %w", err)
	}

	return metrics, nil
}

// GetTeamMetricsHistory retrieves historical metrics for a team
func (r *MetricsRepository) GetTeamMetricsHistory(ctx context.Context, teamID uuid.UUID, limit int) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE team_id = $1
		ORDER BY period_start DESC
		LIMIT $2
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get team metrics history: %w", err)
	}

	return metrics, nil
}

// GetLeaderboard retrieves top users by total commits for a period
func (r *MetricsRepository) GetLeaderboard(ctx context.Context, start, end time.Time, limit int) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id IS NOT NULL
		  AND period_start >= $1 AND period_end <= $2
		ORDER BY total_commits DESC
		LIMIT $3
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return metrics, nil
}

// LeaderboardScoreRow embeds AggregatedMetrics so sqlx StructScan can reach all
// db-tagged fields. Using a named field (Metrics entity.AggregatedMetrics)
// prevents sqlx from descending into the struct.
type LeaderboardScoreRow struct {
	entity.AggregatedMetrics
	Score float64 `db:"score"`
}

// GetLeaderboardByScore retrieves top users by calculated score
func (r *MetricsRepository) GetLeaderboardByScore(ctx context.Context, start, end time.Time, limit int) ([]LeaderboardScoreRow, error) {
	query := `
		SELECT *,
		       (total_commits * 1.0 +
		        total_lines_added * 0.01 +
		        total_prs_merged * 5.0 +
		        total_reviews * 3.0 +
		        total_issues * 2.0) as score
		FROM aggregated_metrics
		WHERE user_id IS NOT NULL
		  AND period_start >= $1 AND period_end <= $2
		ORDER BY score DESC
		LIMIT $3
	`

	var results []LeaderboardScoreRow
	err := r.db.SelectContext(ctx, &results, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard by score: %w", err)
	}

	return results, nil
}

// GetTopTeams retrieves top teams by total commits
func (r *MetricsRepository) GetTopTeams(ctx context.Context, start, end time.Time, limit int) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE team_id IS NOT NULL
		  AND period_start >= $1 AND period_end <= $2
		ORDER BY total_commits DESC
		LIMIT $3
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top teams: %w", err)
	}

	return metrics, nil
}

// GetAllUsersForPeriod retrieves metrics for all users in a period
func (r *MetricsRepository) GetAllUsersForPeriod(ctx context.Context, start, end time.Time) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id IS NOT NULL
		  AND period_start = $1 AND period_end = $2
		ORDER BY total_commits DESC
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get all user metrics: %w", err)
	}

	return metrics, nil
}

// GetAllTeamsForPeriod retrieves metrics for all teams in a period
func (r *MetricsRepository) GetAllTeamsForPeriod(ctx context.Context, start, end time.Time) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE team_id IS NOT NULL
		  AND period_start = $1 AND period_end = $2
		ORDER BY total_commits DESC
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get all team metrics: %w", err)
	}

	return metrics, nil
}

// DeleteOlderThan deletes metrics older than specified period start
func (r *MetricsRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	query := `DELETE FROM aggregated_metrics WHERE period_start < $1`

	result, err := r.db.ExecContext(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old metrics: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// GetUserTrend analyzes trend for a user over multiple periods
func (r *MetricsRepository) GetUserTrend(ctx context.Context, userID uuid.UUID, periods int) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id = $1
		ORDER BY period_start DESC
		LIMIT $2
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, userID, periods)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trend: %w", err)
	}

	return metrics, nil
}

// GetTeamMembersMetrics retrieves metrics for all team members in a period
func (r *MetricsRepository) GetTeamMembersMetrics(ctx context.Context, teamID uuid.UUID, start, end time.Time) ([]entity.AggregatedMetrics, error) {
	query := `
		SELECT am.* FROM aggregated_metrics am
		JOIN team_members tm ON am.user_id = tm.user_id
		WHERE tm.team_id = $1 AND tm.left_at IS NULL
		  AND am.period_start = $2 AND am.period_end = $3
		ORDER BY am.total_commits DESC
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, teamID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members metrics: %w", err)
	}

	return metrics, nil
}

// CompareUsers compares metrics between two users for a period
func (r *MetricsRepository) CompareUsers(ctx context.Context, userID1, userID2 uuid.UUID, start, end time.Time) (map[string]*entity.AggregatedMetrics, error) {
	query := `
		SELECT * FROM aggregated_metrics
		WHERE user_id IN ($1, $2)
		  AND period_start = $3 AND period_end = $4
	`

	var metrics []entity.AggregatedMetrics
	err := r.db.SelectContext(ctx, &metrics, query, userID1, userID2, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to compare users: %w", err)
	}

	result := make(map[string]*entity.AggregatedMetrics)
	for i := range metrics {
		if metrics[i].UserID != nil {
			result[metrics[i].UserID.String()] = &metrics[i]
		}
	}

	return result, nil
}

// GetPeriodsSummary gets summary of all available periods
func (r *MetricsRepository) GetPeriodsSummary(ctx context.Context) ([]struct {
	PeriodStart time.Time `db:"period_start"`
	PeriodEnd   time.Time `db:"period_end"`
	UserCount   int       `db:"user_count"`
	TeamCount   int       `db:"team_count"`
}, error) {
	query := `
		SELECT
			period_start,
			period_end,
			COUNT(DISTINCT user_id) as user_count,
			COUNT(DISTINCT team_id) as team_count
		FROM aggregated_metrics
		GROUP BY period_start, period_end
		ORDER BY period_start DESC
	`

	var summary []struct {
		PeriodStart time.Time `db:"period_start"`
		PeriodEnd   time.Time `db:"period_end"`
		UserCount   int       `db:"user_count"`
		TeamCount   int       `db:"team_count"`
	}

	err := r.db.SelectContext(ctx, &summary, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get periods summary: %w", err)
	}

	return summary, nil
}
