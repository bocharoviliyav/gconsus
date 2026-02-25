package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gconsus/database"
	"gconsus/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ActivityRepository handles git activities database operations
type ActivityRepository struct {
	db *database.DB
}

// NewActivityRepository creates a new ActivityRepository
func NewActivityRepository(db *database.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create creates a new git activity
func (r *ActivityRepository) Create(ctx context.Context, activity *entity.GitActivity) error {
	query := `
		INSERT INTO git_activities (
			id, user_id, provider_id, activity_type, repository_name, repository_owner,
			commit_count, lines_added, lines_deleted, pr_title, pr_url, pr_merged,
			issue_title, issue_url, issue_state, occurred_at, fetched_at, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}

	if activity.FetchedAt.IsZero() {
		activity.FetchedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		activity.ID, activity.UserID, activity.ProviderID, activity.ActivityType,
		activity.RepositoryName, activity.RepositoryOwner, activity.CommitCount,
		activity.LinesAdded, activity.LinesDeleted, activity.PRTitle, activity.PRURL,
		activity.PRMerged, activity.IssueTitle, activity.IssueURL, activity.IssueState,
		activity.OccurredAt, activity.FetchedAt, activity.RawData,
	)

	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	return nil
}

// BatchCreate creates multiple activities efficiently
func (r *ActivityRepository) BatchCreate(ctx context.Context, activities []entity.GitActivity) error {
	if len(activities) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO git_activities (
			id, user_id, provider_id, activity_type, repository_name, repository_owner,
			commit_count, lines_added, lines_deleted, pr_title, pr_url, pr_merged,
			issue_title, issue_url, issue_state, occurred_at, fetched_at, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, activity := range activities {
		if activity.ID == uuid.Nil {
			activity.ID = uuid.New()
		}
		if activity.FetchedAt.IsZero() {
			activity.FetchedAt = now
		}

		_, err = stmt.ExecContext(ctx,
			activity.ID, activity.UserID, activity.ProviderID, activity.ActivityType,
			activity.RepositoryName, activity.RepositoryOwner, activity.CommitCount,
			activity.LinesAdded, activity.LinesDeleted, activity.PRTitle, activity.PRURL,
			activity.PRMerged, activity.IssueTitle, activity.IssueURL, activity.IssueState,
			activity.OccurredAt, activity.FetchedAt, activity.RawData,
		)
		if err != nil {
			return fmt.Errorf("failed to insert activity: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves an activity by ID
func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.GitActivity, error) {
	var activity entity.GitActivity
	query := `SELECT * FROM git_activities WHERE id = $1`

	err := r.db.GetContext(ctx, &activity, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("activity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// GetByUser retrieves activities for a specific user
func (r *ActivityRepository) GetByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]entity.GitActivity, error) {
	query := `
		SELECT * FROM git_activities
		WHERE user_id = $1 AND occurred_at BETWEEN $2 AND $3
		ORDER BY occurred_at DESC
		LIMIT $4 OFFSET $5
	`

	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, userID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activities: %w", err)
	}

	return activities, nil
}

// GetByUserAndType retrieves activities for a user filtered by type
func (r *ActivityRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, activityType entity.ActivityType, startDate, endDate time.Time) ([]entity.GitActivity, error) {
	query := `
		SELECT * FROM git_activities
		WHERE user_id = $1 AND activity_type = $2 AND occurred_at BETWEEN $3 AND $4
		ORDER BY occurred_at DESC
	`

	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, userID, activityType, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by type: %w", err)
	}

	return activities, nil
}

// GetByRepository retrieves activities for a specific repository
func (r *ActivityRepository) GetByRepository(ctx context.Context, owner, name string, startDate, endDate time.Time) ([]entity.GitActivity, error) {
	query := `
		SELECT * FROM git_activities
		WHERE repository_owner = $1 AND repository_name = $2 AND occurred_at BETWEEN $3 AND $4
		ORDER BY occurred_at DESC
	`

	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, owner, name, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository activities: %w", err)
	}

	return activities, nil
}

// GetByTeam retrieves activities for all members of a team
func (r *ActivityRepository) GetByTeam(ctx context.Context, teamID uuid.UUID, startDate, endDate time.Time) ([]entity.GitActivity, error) {
	query := `
		SELECT ga.* FROM git_activities ga
		JOIN team_members tm ON ga.user_id = tm.user_id
		WHERE tm.team_id = $1 AND tm.left_at IS NULL
		  AND ga.occurred_at BETWEEN $2 AND $3
		ORDER BY ga.occurred_at DESC
	`

	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get team activities: %w", err)
	}

	return activities, nil
}

// GetByUsersAndPeriod retrieves activities for multiple users within a date range
func (r *ActivityRepository) GetByUsersAndPeriod(ctx context.Context, userIDs []uuid.UUID, startDate, endDate time.Time) ([]entity.GitActivity, error) {
	if len(userIDs) == 0 {
		return []entity.GitActivity{}, nil
	}

	// Build the query with IN clause for user IDs
	query := `
		SELECT * FROM git_activities
		WHERE user_id = ANY($1) AND occurred_at BETWEEN $2 AND $3
		ORDER BY occurred_at DESC
	`

	pqIDs := make([]string, len(userIDs))
	for i, id := range userIDs {
		pqIDs[i] = id.String()
	}
	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, pq.Array(pqIDs), startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities for users: %w", err)
	}

	return activities, nil
}

// GetByPeriod retrieves all activities within a date range
func (r *ActivityRepository) GetByPeriod(ctx context.Context, startDate, endDate time.Time) ([]entity.GitActivity, error) {
	query := `
		SELECT * FROM git_activities
		WHERE occurred_at BETWEEN $1 AND $2
		ORDER BY occurred_at DESC
	`

	var activities []entity.GitActivity
	err := r.db.SelectContext(ctx, &activities, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by period: %w", err)
	}

	return activities, nil
}

// CountByUser counts activities for a user in a date range
func (r *ActivityRepository) CountByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM git_activities
		WHERE user_id = $1 AND occurred_at BETWEEN $2 AND $3
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to count activities: %w", err)
	}

	return count, nil
}

// GetUserRepositories retrieves all repositories a user contributed to
func (r *ActivityRepository) GetUserRepositories(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]string, error) {
	query := `
		SELECT DISTINCT repository_owner || '/' || repository_name as full_name
		FROM git_activities
		WHERE user_id = $1 AND occurred_at BETWEEN $2 AND $3
		ORDER BY full_name
	`

	var repos []string
	err := r.db.SelectContext(ctx, &repos, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %w", err)
	}

	return repos, nil
}

// GetTopContributors retrieves top contributors by activity count
func (r *ActivityRepository) GetTopContributors(ctx context.Context, startDate, endDate time.Time, limit int) ([]struct {
	UserID        uuid.UUID `db:"user_id"`
	ActivityCount int       `db:"activity_count"`
}, error) {
	query := `
		SELECT user_id, COUNT(*) as activity_count
		FROM git_activities
		WHERE occurred_at BETWEEN $1 AND $2
		GROUP BY user_id
		ORDER BY activity_count DESC
		LIMIT $3
	`

	var contributors []struct {
		UserID        uuid.UUID `db:"user_id"`
		ActivityCount int       `db:"activity_count"`
	}
	err := r.db.SelectContext(ctx, &contributors, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top contributors: %w", err)
	}

	return contributors, nil
}

// DeleteOlderThan deletes activities older than specified date (for data retention)
func (r *ActivityRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	query := `DELETE FROM git_activities WHERE occurred_at < $1`

	result, err := r.db.ExecContext(ctx, query, date)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old activities: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// GetLatestFetchTime retrieves the latest fetch time for a provider
func (r *ActivityRepository) GetLatestFetchTime(ctx context.Context, providerID uuid.UUID) (*time.Time, error) {
	query := `
		SELECT MAX(occurred_at) FROM git_activities
		WHERE provider_id = $1
	`

	var latestTime *time.Time
	err := r.db.GetContext(ctx, &latestTime, query, providerID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest fetch time: %w", err)
	}

	return latestTime, nil
}

// ActivityStats represents aggregated statistics
type ActivityStats struct {
	TotalActivities   int   `db:"total_activities"`
	TotalCommits      int   `db:"total_commits"`
	TotalLinesAdded   int64 `db:"total_lines_added"`
	TotalLinesDeleted int64 `db:"total_lines_deleted"`
	TotalPRs          int   `db:"total_prs"`
	TotalReviews      int   `db:"total_reviews"`
	TotalIssues       int   `db:"total_issues"`
}

// GetStatsForUser calculates statistics for a user
func (r *ActivityRepository) GetStatsForUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*ActivityStats, error) {
	query := `
		SELECT
			COUNT(*) as total_activities,
			SUM(CASE WHEN activity_type = 'commit' THEN commit_count ELSE 0 END) as total_commits,
			SUM(lines_added) as total_lines_added,
			SUM(lines_deleted) as total_lines_deleted,
			COUNT(CASE WHEN activity_type = 'pr' THEN 1 END) as total_prs,
			COUNT(CASE WHEN activity_type = 'review' THEN 1 END) as total_reviews,
			COUNT(CASE WHEN activity_type = 'issue' THEN 1 END) as total_issues
		FROM git_activities
		WHERE user_id = $1 AND occurred_at BETWEEN $2 AND $3
	`

	var stats ActivityStats
	err := r.db.GetContext(ctx, &stats, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}
