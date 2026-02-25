package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gconsus/adapter/vcs"
	"gconsus/entity"

	"github.com/go-playground/validator/v10"
	"github.com/xanzy/go-gitlab"
)

type config struct {
	BaseURL   string `validate:"required,min=1,max=200"`
	AuthToken string `validate:"required,min=1,max=200"`
}

func DefaultConfig() config {
	return config{}
}

// Client provides access to GitLab REST API v4.
// It implements vcs.Client.
type Client struct {
	config       config
	gitlabClient *gitlab.Client
}

func New(cfg config) (Client, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return Client{}, fmt.Errorf("invalid config, %w", err)
	}

	client, err := gitlab.NewClient(cfg.AuthToken, gitlab.WithBaseURL(cfg.BaseURL))
	if err != nil {
		return Client{}, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return Client{config: cfg, gitlabClient: client}, nil
}

// ---------------------------------------------------------------------------
// vcs.Client implementation
// ---------------------------------------------------------------------------

// FetchUserActivities returns commits, MRs, reviews and issues for a user.
func (c *Client) FetchUserActivities(ctx context.Context, username string, from, to time.Time) ([]vcs.Activity, error) {
	userID, err := c.resolveUserID(ctx, username)
	if err != nil {
		return nil, err
	}

	var all []vcs.Activity

	// Commits across all accessible projects
	projects, err := c.listUserProjects(ctx, userID)
	if err != nil {
		slog.Warn("gitlab: list projects", "error", err)
	}
	for _, proj := range projects {
		commits := c.paginateCommits(ctx, proj.ID, proj.PathWithNamespace, from, to)
		all = append(all, commits...)
	}

	// Merge requests (authored)
	mrs := c.paginateMergeRequests(ctx, userID, from, to)
	for _, mr := range mrs {
		merged := mr.State == "merged"
		owner, name := splitProjectRef(mr.References.Full, "!")
		all = append(all, vcs.Activity{
			Type:            vcs.ActivityPR,
			RepositoryOwner: owner,
			RepositoryName:  name,
			Title:           mr.Title,
			URL:             mr.WebURL,
			Merged:          &merged,
			LinesAdded:      atoiSafe(mr.ChangesCount),
			OccurredAt:      safeTime(mr.CreatedAt),
		})
	}

	// Reviews (MRs where user is reviewer, not author)
	reviews := c.paginateReviews(ctx, userID, from, to)
	for _, mr := range reviews {
		owner, name := splitProjectRef(mr.References.Full, "!")
		all = append(all, vcs.Activity{
			Type:            vcs.ActivityReview,
			RepositoryOwner: owner,
			RepositoryName:  name,
			Title:           mr.Title,
			URL:             mr.WebURL,
			OccurredAt:      safeTime(mr.CreatedAt),
		})
	}

	// Issues
	issues := c.paginateIssues(ctx, userID, from, to)
	for _, issue := range issues {
		owner, name := splitProjectRef(issue.References.Full, "#")
		all = append(all, vcs.Activity{
			Type:            vcs.ActivityIssue,
			RepositoryOwner: owner,
			RepositoryName:  name,
			Title:           issue.Title,
			URL:             issue.WebURL,
			State:           issue.State,
			OccurredAt:      safeTime(issue.CreatedAt),
		})
	}

	return all, nil
}

// FetchRepositories returns repositories for a group or the authenticated user.
func (c *Client) FetchRepositories(ctx context.Context, org string) ([]vcs.Repository, error) {
	var repos []vcs.Repository

	if org != "" {
		// Group projects
		page := 1
		for {
			projects, resp, err := c.gitlabClient.Groups.ListGroupProjects(org, &gitlab.ListGroupProjectsOptions{
				ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
			}, gitlab.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("gitlab: list group projects: %w", err)
			}
			for _, p := range projects {
				repos = append(repos, projectToRepo(p))
			}
			if resp.NextPage == 0 {
				break
			}
			page = resp.NextPage
		}
	} else {
		// User's projects
		page := 1
		for {
			projects, resp, err := c.gitlabClient.Projects.ListProjects(&gitlab.ListProjectsOptions{
				Membership:  gitlab.Bool(true),
				ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
			}, gitlab.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("gitlab: list projects: %w", err)
			}
			for _, p := range projects {
				repos = append(repos, projectToRepo(p))
			}
			if resp.NextPage == 0 {
				break
			}
			page = resp.NextPage
		}
	}

	return repos, nil
}

// FetchPullRequestStats returns statistics for a merge request.
func (c *Client) FetchPullRequestStats(ctx context.Context, owner, repo string, prNumber int) (*vcs.PRStats, error) {
	// In GitLab, owner/repo maps to the project path.
	projectPath := owner
	if repo != "" {
		projectPath = owner + "/" + repo
	}

	mr, _, err := c.gitlabClient.MergeRequests.GetMergeRequest(projectPath, prNumber, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("gitlab: get MR !%d: %w", prNumber, err)
	}

	// Count notes as review comments
	notes, _, err := c.gitlabClient.Notes.ListMergeRequestNotes(projectPath, prNumber, &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}, gitlab.WithContext(ctx))
	reviewCount := 0
	if err == nil {
		for _, n := range notes {
			if !n.System {
				reviewCount++
			}
		}
	}

	merged := mr.State == "merged"

	return &vcs.PRStats{
		Number:       mr.IID,
		Title:        mr.Title,
		URL:          mr.WebURL,
		Merged:       merged,
		Additions:    atoiSafe(mr.ChangesCount),
		ReviewCount:  reviewCount,
		OccurredAt:   safeTime(mr.CreatedAt),
	}, nil
}

// FetchUsers returns active users visible to the token.
func (c *Client) FetchUsers(ctx context.Context) ([]vcs.VCSUser, error) {
	var all []vcs.VCSUser
	page := 1
	for {
		users, resp, err := c.gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{
			Active:      gitlab.Ptr(true),
			ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
		}, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("gitlab: list users: %w", err)
		}
		for _, u := range users {
			all = append(all, vcs.VCSUser{
				Username:  u.Username,
				Name:      u.Name,
				Email:     u.Email,
				AvatarURL: u.AvatarURL,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// TestConnection verifies the token by calling /user.
func (c *Client) TestConnection(ctx context.Context) error {
	_, _, err := c.gitlabClient.Users.CurrentUser(gitlab.WithContext(ctx))
	return err
}

// ---------------------------------------------------------------------------
// Legacy method kept for backward compatibility with service.Service
// ---------------------------------------------------------------------------

func (c *Client) UserActivity(
	ctx context.Context, username string, from time.Time, to time.Time,
) (entity.UserActivityInfo, error) {
	userID, err := c.resolveUserID(ctx, username)
	if err != nil {
		return entity.UserActivityInfo{}, err
	}

	result := entity.UserActivityInfo{
		Login: username,
	}

	projects, err := c.listUserProjects(ctx, userID)
	if err != nil {
		slog.Warn("failed to list projects", "error", err)
	}
	commitCount := 0
	for _, proj := range projects {
		commits := c.paginateCommits(ctx, proj.ID, proj.PathWithNamespace, from, to)
		commitCount += len(commits)
	}
	result.ContributionsCollection.TotalCommitContributions = commitCount

	mrs := c.paginateMergeRequests(ctx, userID, from, to)
	result.ContributionsCollection.TotalPullRequestContributions = len(mrs)

	issues := c.paginateIssues(ctx, userID, from, to)
	result.ContributionsCollection.TotalIssueContributions = len(issues)

	reviews := c.paginateReviews(ctx, userID, from, to)
	result.ContributionsCollection.TotalPullRequestReviewContributions = len(reviews)

	return result, nil
}

// ---------------------------------------------------------------------------
// Internal helpers with pagination
// ---------------------------------------------------------------------------

// resolveUserID looks up a GitLab user by username.
func (c *Client) resolveUserID(ctx context.Context, username string) (int, error) {
	users, _, err := c.gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.String(username),
	}, gitlab.WithContext(ctx))
	if err != nil {
		slog.Error("failed to get GitLab user", "error", err, "username", username)
		return 0, err
	}
	if len(users) == 0 {
		return 0, fmt.Errorf("gitlab: user not found: %s", username)
	}
	return users[0].ID, nil
}

// listUserProjects returns all projects a user has access to (paginated).
func (c *Client) listUserProjects(ctx context.Context, userID int) ([]*gitlab.Project, error) {
	var all []*gitlab.Project
	page := 1
	for {
		projects, resp, err := c.gitlabClient.Projects.ListUserProjects(userID, &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
		}, gitlab.WithContext(ctx))
		if err != nil {
			return all, err
		}
		all = append(all, projects...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// paginateCommits returns commit activities across all pages for a project.
func (c *Client) paginateCommits(ctx context.Context, projectID int, projectPath string, from, to time.Time) []vcs.Activity {
	var activities []vcs.Activity
	page := 1
	for {
		opt := &gitlab.ListCommitsOptions{
			Since:       &from,
			Until:       &to,
			ListOptions: gitlab.ListOptions{PerPage: 100, Page: page},
		}
		commits, resp, err := c.gitlabClient.Commits.ListCommits(projectID, opt, gitlab.WithContext(ctx))
		if err != nil {
			slog.Warn("gitlab: list commits", "project", projectPath, "error", err)
			break
		}
		for _, cm := range commits {
			cOwner, cName := splitProjectPath(projectPath)
			a := vcs.Activity{
				Type:            vcs.ActivityCommit,
				RepositoryOwner: cOwner,
				RepositoryName:  cName,
				Title:           cm.Title,
				CommitCount:     1,
				OccurredAt:      safeTime(cm.CreatedAt),
			}
			if cm.Stats != nil {
				a.LinesAdded = cm.Stats.Additions
				a.LinesDeleted = cm.Stats.Deletions
			}
			activities = append(activities, a)
		}
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return activities
}

// paginateMergeRequests returns all merge requests authored by user.
func (c *Client) paginateMergeRequests(ctx context.Context, userID int, from, to time.Time) []*gitlab.MergeRequest {
	var all []*gitlab.MergeRequest
	page := 1
	for {
		opt := &gitlab.ListMergeRequestsOptions{
			AuthorID:      gitlab.Int(userID),
			CreatedAfter:  &from,
			CreatedBefore: &to,
			ListOptions:   gitlab.ListOptions{PerPage: 100, Page: page},
		}
		mrs, resp, err := c.gitlabClient.MergeRequests.ListMergeRequests(opt, gitlab.WithContext(ctx))
		if err != nil {
			slog.Warn("gitlab: list MRs", "error", err)
			break
		}
		all = append(all, mrs...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all
}

// paginateReviews returns MRs where the user is a reviewer (not author).
func (c *Client) paginateReviews(ctx context.Context, userID int, from, to time.Time) []*gitlab.MergeRequest {
	var all []*gitlab.MergeRequest
	page := 1
	for {
		opt := &gitlab.ListMergeRequestsOptions{
			ReviewerID:    gitlab.ReviewerID(userID),
			CreatedAfter:  &from,
			CreatedBefore: &to,
			ListOptions:   gitlab.ListOptions{PerPage: 100, Page: page},
		}
		mrs, resp, err := c.gitlabClient.MergeRequests.ListMergeRequests(opt, gitlab.WithContext(ctx))
		if err != nil {
			slog.Warn("gitlab: list reviews", "error", err)
			break
		}
		all = append(all, mrs...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all
}

// paginateIssues returns all issues authored by user.
func (c *Client) paginateIssues(ctx context.Context, userID int, from, to time.Time) []*gitlab.Issue {
	var all []*gitlab.Issue
	page := 1
	for {
		opt := &gitlab.ListIssuesOptions{
			AuthorID:      gitlab.Int(userID),
			CreatedAfter:  &from,
			CreatedBefore: &to,
			ListOptions:   gitlab.ListOptions{PerPage: 100, Page: page},
		}
		issues, resp, err := c.gitlabClient.Issues.ListIssues(opt, gitlab.WithContext(ctx))
		if err != nil {
			slog.Warn("gitlab: list issues", "error", err)
			break
		}
		all = append(all, issues...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all
}

// projectToRepo converts a GitLab Project to vcs.Repository.
func projectToRepo(p *gitlab.Project) vcs.Repository {
	var owner string
	if p.Namespace != nil {
		owner = p.Namespace.FullPath
	}
	return vcs.Repository{
		Owner:       owner,
		Name:        p.Name,
		HTMLURL:     p.WebURL,
		Description: p.Description,
		Stars:       p.StarCount,
		Forks:       p.ForksCount,
		CreatedAt:   safeTime(p.CreatedAt),
		UpdatedAt:   safeTime(p.LastActivityAt),
	}
}

// safeTime dereferences a *time.Time pointer, returning zero value if nil.
func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// atoiSafe converts a string to int, returning 0 on failure.
func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// splitProjectPath splits "group/project-name" into ("group", "project-name").
func splitProjectPath(path string) (owner, name string) {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[:idx], path[idx+1:]
	}
	return "", path
}

// splitProjectRef strips the ref suffix (!N or #N) from a GitLab reference
// like "devteam/frontend-app!25" and splits the project path into owner/name.
var refSuffixRe = regexp.MustCompile(`[!#]\d+$`)

func splitProjectRef(ref, _ string) (owner, name string) {
	path := refSuffixRe.ReplaceAllString(ref, "")
	return splitProjectPath(path)
}
