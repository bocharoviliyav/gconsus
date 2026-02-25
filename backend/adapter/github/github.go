package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gconsus/adapter/vcs"
	"gconsus/entity"

	"github.com/go-playground/validator/v10"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// config holds GitHub connection parameters.
type config struct {
	BaseURL   string `validate:"required,min=1,max=200"`
	AuthToken string `validate:"required,min=1,max=200"`
}

func DefaultConfig() config {
	return config{}
}

// Client provides access to GitHub via both GraphQL and REST APIs.
// It implements vcs.Client.
type Client struct {
	config        config
	graphQLClient *githubv4.Client
	httpClient    *http.Client
	restBaseURL   string
}

// New creates a GitHub client that satisfies vcs.Client.
func New(cfg config) (Client, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return Client{}, fmt.Errorf("invalid config, %w", err)
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.AuthToken})
	httpClient := oauth2.NewClient(context.Background(), src)
	gqlClient := githubv4.NewEnterpriseClient(cfg.BaseURL, httpClient)

	// Derive REST base URL from GraphQL URL.
	// Enterprise: https://host/api/graphql -> https://host/api/v3
	// Public:     https://api.github.com/graphql -> https://api.github.com
	restBase := strings.TrimSuffix(cfg.BaseURL, "/graphql")
	if strings.HasSuffix(restBase, "/api") {
		restBase = restBase + "/v3"
	}

	return Client{
		config:        cfg,
		graphQLClient: gqlClient,
		httpClient:    httpClient,
		restBaseURL:   restBase,
	}, nil
}

// ---------------------------------------------------------------------------
// vcs.Client implementation
// ---------------------------------------------------------------------------

// FetchUserActivities returns commits, PRs, reviews and issues for a user.
func (c *Client) FetchUserActivities(ctx context.Context, username string, from, to time.Time) ([]vcs.Activity, error) {
	var all []vcs.Activity

	// 1. Contributions via GraphQL
	gqlActivities, err := c.fetchContributions(ctx, username, from, to)
	if err != nil {
		slog.Warn("github: GraphQL contributions failed", "error", err)
	} else {
		all = append(all, gqlActivities...)
	}

	// 2. Enrich PRs with additions/deletions via REST
	for i := range all {
		if all[i].Type == vcs.ActivityPR && all[i].URL != "" {
			stats, err := c.fetchPRStatsFromURL(ctx, all[i].URL)
			if err == nil && stats != nil {
				all[i].LinesAdded = stats.Additions
				all[i].LinesDeleted = stats.Deletions
			}
		}
	}

	return all, nil
}

// FetchRepositories lists organisation or user repositories.
func (c *Client) FetchRepositories(ctx context.Context, org string) ([]vcs.Repository, error) {
	var repos []vcs.Repository
	path := "/user/repos?per_page=100&type=all"
	if org != "" {
		path = fmt.Sprintf("/orgs/%s/repos?per_page=100&type=all", org)
	}

	err := c.paginateREST(ctx, path, func(body []byte) error {
		var page []struct {
			Name        string `json:"name"`
			HTMLURL     string `json:"html_url"`
			Description string `json:"description"`
			Language    string `json:"language"`
			Stars       int    `json:"stargazers_count"`
			Forks       int    `json:"forks_count"`
			Owner       struct {
				Login string `json:"login"`
			} `json:"owner"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, r := range page {
			repos = append(repos, vcs.Repository{
				Owner:       r.Owner.Login,
				Name:        r.Name,
				HTMLURL:     r.HTMLURL,
				Description: r.Description,
				Language:    r.Language,
				Stars:       r.Stars,
				Forks:       r.Forks,
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
			})
		}
		return nil
	})
	return repos, err
}

// FetchPullRequestStats returns stats for a single pull request via REST.
func (c *Client) FetchPullRequestStats(ctx context.Context, owner, repo string, prNumber int) (*vcs.PRStats, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, prNumber)
	body, err := c.doREST(ctx, path)
	if err != nil {
		return nil, err
	}

	var pr struct {
		Number       int       `json:"number"`
		Title        string    `json:"title"`
		HTMLURL      string    `json:"html_url"`
		Merged       bool      `json:"merged"`
		Additions    int       `json:"additions"`
		Deletions    int       `json:"deletions"`
		ChangedFiles int       `json:"changed_files"`
		CreatedAt    time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("github: parse PR: %w", err)
	}

	// Fetch review count
	reviewPath := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews?per_page=100", owner, repo, prNumber)
	reviewBody, err := c.doREST(ctx, reviewPath)
	reviewCount := 0
	if err == nil {
		var reviews []json.RawMessage
		if json.Unmarshal(reviewBody, &reviews) == nil {
			reviewCount = len(reviews)
		}
	}

	return &vcs.PRStats{
		Number:       pr.Number,
		Title:        pr.Title,
		URL:          pr.HTMLURL,
		Merged:       pr.Merged,
		Additions:    pr.Additions,
		Deletions:    pr.Deletions,
		ChangedFiles: pr.ChangedFiles,
		ReviewCount:  reviewCount,
		OccurredAt:   pr.CreatedAt,
	}, nil
}

// FetchUsers returns users by discovering org members from accessible repos.
func (c *Client) FetchUsers(ctx context.Context) ([]vcs.VCSUser, error) {
	// Discover orgs from accessible repos.
	repos, err := c.FetchRepositories(ctx, "")
	if err != nil {
		return nil, err
	}
	orgs := map[string]struct{}{}
	for _, r := range repos {
		if r.Owner != "" {
			orgs[r.Owner] = struct{}{}
		}
	}

	seen := map[string]struct{}{}
	var all []vcs.VCSUser

	for org := range orgs {
		path := fmt.Sprintf("/orgs/%s/members?per_page=100", org)
		err := c.paginateREST(ctx, path, func(body []byte) error {
			var members []struct {
				Login     string `json:"login"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				AvatarURL string `json:"avatar_url"`
			}
			if err := json.Unmarshal(body, &members); err != nil {
				return err
			}
			for _, m := range members {
				if _, ok := seen[m.Login]; ok {
					continue
				}
				seen[m.Login] = struct{}{}
				all = append(all, vcs.VCSUser{
					Username:  m.Login,
					Name:      m.Name,
					Email:     m.Email,
					AvatarURL: m.AvatarURL,
				})
			}
			return nil
		})
		if err != nil {
			slog.Warn("github: list org members", "org", org, "error", err)
		}
	}

	// Fallback: if no org members found, return authenticated user.
	if len(all) == 0 {
		body, err := c.doREST(ctx, "/user")
		if err != nil {
			return nil, err
		}
		var u struct {
			Login     string `json:"login"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		}
		if json.Unmarshal(body, &u) == nil && u.Login != "" {
			all = append(all, vcs.VCSUser{
				Username:  u.Login,
				Name:      u.Name,
				Email:     u.Email,
				AvatarURL: u.AvatarURL,
			})
		}
	}

	return all, nil
}

// TestConnection verifies the token by calling /user.
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.doREST(ctx, "/user")
	return err
}

// ---------------------------------------------------------------------------
// Legacy method kept for backward compatibility with service.Service
// ---------------------------------------------------------------------------

func (c *Client) UserActivity(
	ctx context.Context, login string, from time.Time, to time.Time,
) (entity.UserActivityQuery, error) {
	variables := map[string]any{
		"login": githubv4.String(login),
		"from":  githubv4.DateTime{Time: from},
		"to":    githubv4.DateTime{Time: to},
	}

	var query entity.UserActivityQuery
	if err := c.graphQLClient.Query(ctx, &query, variables); err != nil {
		slog.Error(err.Error())
		return query, err
	}
	return query, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// fetchContributions uses the GraphQL API to get the contributionsCollection.
func (c *Client) fetchContributions(ctx context.Context, login string, from, to time.Time) ([]vcs.Activity, error) {
	variables := map[string]any{
		"login": githubv4.String(login),
		"from":  githubv4.DateTime{Time: from},
		"to":    githubv4.DateTime{Time: to},
	}

	var query entity.UserActivityQuery
	if err := c.graphQLClient.Query(ctx, &query, variables); err != nil {
		return nil, fmt.Errorf("github graphql: %w", err)
	}

	var activities []vcs.Activity
	cc := query.User.ContributionsCollection

	// Commits by repository
	for _, repo := range cc.CommitContributionsByRepository {
		for _, node := range repo.Contributions.Nodes {
			activities = append(activities, vcs.Activity{
				Type:            vcs.ActivityCommit,
				RepositoryOwner: repo.Repository.Owner.Login,
				RepositoryName:  repo.Repository.Name,
				CommitCount:     node.CommitCount,
				OccurredAt:      node.OccurredAt,
			})
		}
	}

	// Pull requests by repository
	for _, repo := range cc.PullRequestContributionsByRepository {
		for _, node := range repo.Contributions.Nodes {
			merged := node.PullRequest.Merged
			activities = append(activities, vcs.Activity{
				Type:            vcs.ActivityPR,
				RepositoryOwner: repo.Repository.Owner.Login,
				RepositoryName:  repo.Repository.Name,
				Title:           node.PullRequest.Title,
				URL:             node.PullRequest.Url,
				Merged:          &merged,
				OccurredAt:      node.OccurredAt,
			})
		}
	}

	// Issues by repository
	for _, repo := range cc.IssueContributionsByRepository {
		for _, node := range repo.Contributions.Nodes {
			activities = append(activities, vcs.Activity{
				Type:            vcs.ActivityIssue,
				RepositoryOwner: repo.Repository.Owner.Login,
				RepositoryName:  repo.Repository.Name,
				Title:           node.Issue.Title,
				URL:             node.Issue.Url,
				State:           node.Issue.State,
				OccurredAt:      node.OccurredAt,
			})
		}
	}

	// Reviews
	for _, node := range cc.PullRequestReviewContributions.Nodes {
		activities = append(activities, vcs.Activity{
			Type:            vcs.ActivityReview,
			RepositoryOwner: node.PullRequest.Repository.Owner.Login,
			RepositoryName:  node.PullRequest.Repository.Name,
			Title:           node.PullRequest.Title,
			URL:             node.PullRequest.Url,
			OccurredAt:      node.OccurredAt,
		})
	}

	return activities, nil
}

// fetchPRStatsFromURL extracts owner/repo/number from a GitHub PR HTML URL.
func (c *Client) fetchPRStatsFromURL(ctx context.Context, htmlURL string) (*vcs.PRStats, error) {
	re := regexp.MustCompile(`/([^/]+)/([^/]+)/pull/(\d+)`)
	matches := re.FindStringSubmatch(htmlURL)
	if len(matches) < 4 {
		return nil, fmt.Errorf("cannot parse PR URL: %s", htmlURL)
	}
	var prNumber int
	fmt.Sscanf(matches[3], "%d", &prNumber)
	return c.FetchPullRequestStats(ctx, matches[1], matches[2], prNumber)
}

// doREST performs a single authenticated GET request against the REST API.
func (c *Client) doREST(ctx context.Context, path string) ([]byte, error) {
	url := c.restBaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github REST %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		trunc := len(body)
		if trunc > 200 {
			trunc = 200
		}
		return nil, fmt.Errorf("github REST %s: status %d: %s", path, resp.StatusCode, string(body[:trunc]))
	}
	return body, nil
}

// paginateREST follows Link header pagination, calling fn for each page body.
func (c *Client) paginateREST(ctx context.Context, path string, fn func([]byte) error) error {
	nextURL := c.restBaseURL + path

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("github REST paginate: %w", err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("github REST paginate: status %d", resp.StatusCode)
		}

		if err := fn(body); err != nil {
			return err
		}

		nextURL = parseLinkNext(resp.Header.Get("Link"))
	}

	return nil
}

// parseLinkNext extracts the "next" URL from GitHub Link header.
func parseLinkNext(header string) string {
	if header == "" {
		return ""
	}
	re := regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
	matches := re.FindStringSubmatch(header)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
