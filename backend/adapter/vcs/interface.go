package vcs

import (
	"context"
	"time"
)

// ProviderType identifies the VCS provider.
type ProviderType string

const (
	ProviderGitHub ProviderType = "github"
	ProviderGitLab ProviderType = "gitlab"
)

// Activity represents a single developer activity fetched from a VCS provider.
type Activity struct {
	Type            ActivityType
	RepositoryOwner string
	RepositoryName  string
	Title           string
	URL             string
	CommitCount     int
	LinesAdded      int
	LinesDeleted    int
	Merged          *bool
	State           string // open, closed, merged
	OccurredAt      time.Time
	RawJSON         []byte // full API response for raw_data column
}

// ActivityType mirrors entity.ActivityType but lives in the adapter layer.
type ActivityType string

const (
	ActivityCommit ActivityType = "commit"
	ActivityPR     ActivityType = "pr"
	ActivityReview ActivityType = "review"
	ActivityIssue  ActivityType = "issue"
)

// VCSUser represents a user discovered from a VCS provider.
type VCSUser struct {
	Username  string
	Name      string
	Email     string
	AvatarURL string
}

// Repository represents a repository returned by the VCS provider.
type Repository struct {
	Owner       string
	Name        string
	HTMLURL     string
	Description string
	Language    string
	Stars       int
	Forks       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PRStats holds detailed pull request statistics.
type PRStats struct {
	Number       int
	Title        string
	URL          string
	Merged       bool
	Additions    int
	Deletions    int
	ChangedFiles int
	ReviewCount  int
	OccurredAt   time.Time
}

// Client is the unified interface every VCS adapter must implement.
// The sync service works exclusively through this interface.
type Client interface {
	// FetchUserActivities returns all activities (commits, PRs, reviews, issues)
	// for a given username within the date range.
	// Implementations must handle pagination internally and return the full set.
	FetchUserActivities(ctx context.Context, username string, from, to time.Time) ([]Activity, error)

	// FetchRepositories returns repositories accessible to the configured token.
	// org may be empty -- in that case return all repositories the token has access to.
	FetchRepositories(ctx context.Context, org string) ([]Repository, error)

	// FetchPullRequestStats returns detailed statistics for a single PR.
	FetchPullRequestStats(ctx context.Context, owner, repo string, prNumber int) (*PRStats, error)

	// FetchUsers returns users visible to the configured token.
	FetchUsers(ctx context.Context) ([]VCSUser, error)

	// TestConnection verifies that the configured credentials are valid.
	TestConnection(ctx context.Context) error
}
