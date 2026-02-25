package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a developer/employee
type User struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	FirstName  string    `json:"firstName" db:"first_name"`
	LastName   string    `json:"lastName" db:"last_name"`
	Patronymic *string   `json:"patronymic,omitempty" db:"patronymic"`
	Email      *string   `json:"email,omitempty" db:"email"`
	PhotoURL   *string   `json:"photoUrl,omitempty" db:"photo_url"`
	Position   *string   `json:"position,omitempty" db:"position"`
	IsActive   bool      `json:"isActive" db:"is_active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// Team represents a developer team
type Team struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	ManagerID   *uuid.UUID `json:"managerId,omitempty" db:"manager_id"`
	IsActive    bool       `json:"isActive" db:"is_active"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

// TeamWithMembers extends Team with member information
type TeamWithMembers struct {
	Team
	Manager *User        `json:"manager,omitempty"`
	Members []TeamMember `json:"members"`
}

// TeamMember represents team membership
type TeamMember struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	TeamID   uuid.UUID  `json:"teamId" db:"team_id"`
	UserID   uuid.UUID  `json:"userId" db:"user_id"`
	Role     string     `json:"role" db:"role"`
	JoinedAt time.Time  `json:"joinedAt" db:"joined_at"`
	LeftAt   *time.Time `json:"leftAt,omitempty" db:"left_at"`
	User     *User      `json:"user,omitempty" db:"-"`
}

// VCSProvider represents a VCS system configuration
type VCSProvider struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // github, gitlab
	BaseURL   string    `json:"baseUrl" db:"base_url"`
	AuthToken string    `json:"-" db:"auth_token"` // Never expose in JSON
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// ActivityType represents the type of git activity
type ActivityType string

const (
	ActivityTypeCommit ActivityType = "commit"
	ActivityTypePR     ActivityType = "pr"
	ActivityTypeIssue  ActivityType = "issue"
	ActivityTypeReview ActivityType = "review"
)

// GitActivity represents a single git activity record
type GitActivity struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	UserID          uuid.UUID    `json:"userId" db:"user_id"`
	ProviderID      uuid.UUID    `json:"providerId" db:"provider_id"`
	ActivityType    ActivityType `json:"activityType" db:"activity_type"`
	RepositoryName  string       `json:"repositoryName" db:"repository_name"`
	RepositoryOwner string       `json:"repositoryOwner" db:"repository_owner"`
	CommitCount     int          `json:"commitCount" db:"commit_count"`
	LinesAdded      int          `json:"linesAdded" db:"lines_added"`
	LinesDeleted    int          `json:"linesDeleted" db:"lines_deleted"`
	PRTitle         *string      `json:"prTitle,omitempty" db:"pr_title"`
	PRURL           *string      `json:"prUrl,omitempty" db:"pr_url"`
	PRMerged        *bool        `json:"prMerged,omitempty" db:"pr_merged"`
	IssueTitle      *string      `json:"issueTitle,omitempty" db:"issue_title"`
	IssueURL        *string      `json:"issueUrl,omitempty" db:"issue_url"`
	IssueState      *string      `json:"issueState,omitempty" db:"issue_state"`
	OccurredAt      time.Time    `json:"occurredAt" db:"occurred_at"`
	FetchedAt       time.Time    `json:"fetchedAt" db:"fetched_at"`
	RawData         JSONB        `json:"rawData,omitempty" db:"raw_data"`
}

// AggregatedMetrics represents pre-calculated metrics
type AggregatedMetrics struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            *uuid.UUID `json:"userId,omitempty" db:"user_id"`
	TeamID            *uuid.UUID `json:"teamId,omitempty" db:"team_id"`
	PeriodStart       time.Time  `json:"periodStart" db:"period_start"`
	PeriodEnd         time.Time  `json:"periodEnd" db:"period_end"`
	TotalCommits      int        `json:"totalCommits" db:"total_commits"`
	TotalLinesAdded   int64      `json:"totalLinesAdded" db:"total_lines_added"`
	TotalLinesDeleted int64      `json:"totalLinesDeleted" db:"total_lines_deleted"`
	TotalPRs          int        `json:"totalPrs" db:"total_prs"`
	TotalPRsMerged    int        `json:"totalPrsMerged" db:"total_prs_merged"`
	TotalReviews      int        `json:"totalReviews" db:"total_reviews"`
	TotalIssues       int        `json:"totalIssues" db:"total_issues"`
	RepositoriesCount int        `json:"repositoriesCount" db:"repositories_count"`
	TopRepositories   JSONB      `json:"topRepositories,omitempty" db:"top_repositories"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
}

// RepositoryStats represents statistics for a repository
type RepositoryStats struct {
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	Commits      int    `json:"commits"`
	LinesAdded   int64  `json:"linesAdded"`
	LinesDeleted int64  `json:"linesDeleted"`
	PullRequests int    `json:"pullRequests"`
	Reviews      int    `json:"reviews"`
}

// Configuration represents a system configuration
type Configuration struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Key         string     `json:"key" db:"key"`
	Value       JSONValue  `json:"value" db:"value"`
	Description *string    `json:"description,omitempty" db:"description"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	UpdatedBy   *uuid.UUID `json:"updatedBy,omitempty" db:"updated_by"`
}

// SyncType represents the type of synchronization
type SyncType string

const (
	SyncTypeEmployees     SyncType = "employees"
	SyncTypeGitActivities SyncType = "git_activities"
	SyncTypeAggregation   SyncType = "aggregation"
)

// SyncStatus represents the status of synchronization
type SyncStatus string

const (
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
)

// SyncHistory represents a synchronization job record
type SyncHistory struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	ProviderID       *uuid.UUID `json:"providerId,omitempty" db:"provider_id"`
	SyncType         SyncType   `json:"syncType" db:"sync_type"`
	Status           SyncStatus `json:"status" db:"status"`
	UsersSynced      int        `json:"usersSynced" db:"users_synced"`
	ActivitiesSynced int        `json:"activitiesSynced" db:"activities_synced"`
	StartedAt        time.Time  `json:"startedAt" db:"started_at"`
	CompletedAt      *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	ErrorMessage     *string    `json:"errorMessage,omitempty" db:"error_message"`
}

// JSONB is a custom type for PostgreSQL JSONB columns that hold JSON objects.
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// JSONValue stores any JSON value (string, number, object, array, bool, null).
// Unlike JSONB it is not restricted to JSON objects.
type JSONValue []byte

func (j JSONValue) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONValue) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	cp := make(JSONValue, len(b))
	copy(cp, b)
	*j = cp
	return nil
}

func (j JSONValue) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONValue) UnmarshalJSON(data []byte) error {
	cp := make(JSONValue, len(data))
	copy(cp, data)
	*j = cp
	return nil
}

// LeaderboardEntry represents a user's position in the leaderboard
type LeaderboardEntry struct {
	Rank              int     `json:"rank"`
	User              User    `json:"user"`
	TotalCommits      int     `json:"totalCommits"`
	TotalLinesAdded   int64   `json:"totalLinesAdded"`
	TotalLinesDeleted int64   `json:"totalLinesDeleted"`
	TotalPRs          int     `json:"totalPrs"`
	TotalReviews      int     `json:"totalReviews"`
	Score             float64 `json:"score"` // Calculated score
}

// TeamLeaderboardEntry represents a team's position
type TeamLeaderboardEntry struct {
	Rank              int     `json:"rank"`
	Team              Team    `json:"team"`
	MembersCount      int     `json:"membersCount"`
	TotalCommits      int     `json:"totalCommits"`
	TotalLinesAdded   int64   `json:"totalLinesAdded"`
	TotalLinesDeleted int64   `json:"totalLinesDeleted"`
	TotalPRs          int     `json:"totalPrs"`
	TotalReviews      int     `json:"totalReviews"`
	Score             float64 `json:"score"`
}

// CreateTeamRequest represents a request to create a team
type CreateTeamRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=255"`
	Description *string    `json:"description,omitempty"`
	ManagerID   *uuid.UUID `json:"lead_id,omitempty"`
}

// UpdateTeamRequest represents a request to update a team
type UpdateTeamRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string    `json:"description,omitempty"`
	ManagerID   *uuid.UUID `json:"lead_id,omitempty"`
	IsActive    *bool      `json:"isActive,omitempty"`
}

// AddTeamMemberRequest represents a request to add a member to a team
type AddTeamMemberRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=developer lead architect qa analyst devops sre"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username    string  `json:"username" validate:"required,min=2,max=100"`
	FirstName  string  `json:"firstName" validate:"required,min=1,max=100"`
	LastName   string  `json:"lastName" validate:"required,min=1,max=100"`
	Patronymic *string `json:"patronymic,omitempty" validate:"omitempty,max=100"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	PhotoURL   *string `json:"photoUrl,omitempty" validate:"omitempty,url"`
	Position   *string `json:"position,omitempty" validate:"omitempty,max=255"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName  *string `json:"firstName,omitempty" validate:"omitempty,min=1,max=100"`
	LastName   *string `json:"lastName,omitempty" validate:"omitempty,min=1,max=100"`
	Patronymic *string `json:"patronymic,omitempty" validate:"omitempty,max=100"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	PhotoURL   *string `json:"photoUrl,omitempty" validate:"omitempty,url"`
	Position   *string `json:"position,omitempty" validate:"omitempty,max=255"`
	IsActive   *bool   `json:"isActive,omitempty"`
}

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	StartDate     time.Time      `json:"startDate"`
	EndDate       time.Time      `json:"endDate"`
	UserIDs       []uuid.UUID    `json:"userIds,omitempty"`
	TeamIDs       []uuid.UUID    `json:"teamIds,omitempty"`
	ProviderIDs   []uuid.UUID    `json:"providerIds,omitempty"`
	ActivityTypes []ActivityType `json:"activityTypes,omitempty"`
	Repositories  []string       `json:"repositories,omitempty"`
}
