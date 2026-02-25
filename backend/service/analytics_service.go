package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gconsus/entity"
	"gconsus/repository"

	"github.com/google/uuid"
)

// AnalyticsService handles analytics and metrics retrieval
type AnalyticsService struct {
	metricsRepo  *repository.MetricsRepository
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
	teamRepo     *repository.TeamRepository
	aggService   *AggregationService
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	metricsRepo *repository.MetricsRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	teamRepo *repository.TeamRepository,
	aggService *AggregationService,
) *AnalyticsService {
	return &AnalyticsService{
		metricsRepo:  metricsRepo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		aggService:   aggService,
	}
}

// GetLeaderboard retrieves the global leaderboard
func (s *AnalyticsService) GetLeaderboard(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.LeaderboardEntry, error) {
	// Get metrics with scores
	metricsWithScores, err := s.metricsRepo.GetLeaderboardByScore(ctx, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Collect user IDs
	var userIDs []uuid.UUID
	for _, m := range metricsWithScores {
		if m.UserID != nil {
			userIDs = append(userIDs, *m.UserID)
		}
	}

	// Get user details
	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Create user map for quick lookup
	userMap := make(map[uuid.UUID]entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Build leaderboard entries
	leaderboard := make([]entity.LeaderboardEntry, 0, len(metricsWithScores))
	for i, m := range metricsWithScores {
		if m.UserID == nil {
			continue
		}

		user, ok := userMap[*m.UserID]
		if !ok {
			continue
		}

		entry := entity.LeaderboardEntry{
			Rank:              i + 1,
			User:              user,
			TotalCommits:      m.TotalCommits,
			TotalLinesAdded:   m.TotalLinesAdded,
			TotalLinesDeleted: m.TotalLinesDeleted,
			TotalPRs:          m.TotalPRs,
			TotalReviews:      m.TotalReviews,
			Score:             m.Score,
		}

		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// GetTeamLeaderboard retrieves the team leaderboard
func (s *AnalyticsService) GetTeamLeaderboard(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TeamLeaderboardEntry, error) {
	// Get team metrics
	teamMetrics, err := s.metricsRepo.GetTopTeams(ctx, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get team leaderboard: %w", err)
	}

	// Collect team IDs
	var teamIDs []uuid.UUID
	for _, m := range teamMetrics {
		if m.TeamID != nil {
			teamIDs = append(teamIDs, *m.TeamID)
		}
	}

	// Get team details
	var teams []entity.Team
	for _, teamID := range teamIDs {
		team, err := s.teamRepo.GetByID(ctx, teamID)
		if err != nil {
			continue
		}
		teams = append(teams, *team)
	}

	// Create team map
	teamMap := make(map[uuid.UUID]entity.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Build leaderboard entries
	leaderboard := make([]entity.TeamLeaderboardEntry, 0, len(teamMetrics))
	for i, m := range teamMetrics {
		if m.TeamID == nil {
			continue
		}

		team, ok := teamMap[*m.TeamID]
		if !ok {
			continue
		}

		// Get member count
		members, _ := s.teamRepo.GetMembers(ctx, team.ID)
		membersCount := len(members)

		score := s.aggService.CalculateScore(&m)

		entry := entity.TeamLeaderboardEntry{
			Rank:              i + 1,
			Team:              team,
			MembersCount:      membersCount,
			TotalCommits:      m.TotalCommits,
			TotalLinesAdded:   m.TotalLinesAdded,
			TotalLinesDeleted: m.TotalLinesDeleted,
			TotalPRs:          m.TotalPRs,
			TotalReviews:      m.TotalReviews,
			Score:             score,
		}

		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// GetUserAnalytics retrieves detailed analytics for a user
func (s *AnalyticsService) GetUserAnalytics(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, repoPage, repoPageSize int) (map[string]interface{}, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get aggregated metrics
	metrics, err := s.metricsRepo.GetByUserAndPeriod(ctx, userID, startDate, endDate)
	if err != nil {
		// If no metrics exist, try to aggregate on-the-fly
		if err := s.aggService.AggregateUserMetrics(ctx, userID, startDate, endDate); err != nil {
			return nil, fmt.Errorf("failed to get user metrics: %w", err)
		}
		metrics, err = s.metricsRepo.GetByUserAndPeriod(ctx, userID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get user metrics: %w", err)
		}
	}

	// Get user's teams
	teams, _ := s.teamRepo.GetUserTeams(ctx, userID)

	// Get user repository contributions from DB
	repoContributions, err := s.getUserRepoContributions(ctx, userID, startDate, endDate)
	if err != nil {
		repoContributions = []map[string]interface{}{}
	}

	// Calculate score
	score := s.aggService.CalculateScore(metrics)

	// Get activity timeline data
	activityTimeline, err := s.getUserActivityTimeline(ctx, userID, startDate, endDate)
	if err != nil {
		// If we can't get timeline data, continue with empty array
		activityTimeline = []map[string]interface{}{}
	}

	// Calculate repository pagination
	totalRepos := len(repoContributions)
	totalRepoPages := (totalRepos + repoPageSize - 1) / repoPageSize
	repoOffset := (repoPage - 1) * repoPageSize
	repoEnd := repoOffset + repoPageSize
	if repoEnd > totalRepos {
		repoEnd = totalRepos
	}

	paginatedRepos := repoContributions
	if repoOffset < totalRepos {
		paginatedRepos = repoContributions[repoOffset:repoEnd]
	} else {
		paginatedRepos = []map[string]interface{}{}
	}

	// Flatten teams to name strings
	teamNames := make([]string, 0, len(teams))
	for _, tm := range teams {
		teamNames = append(teamNames, tm.Name)
	}

	// Derive user email
	var userEmail string
	if user.Email != nil {
		userEmail = *user.Email
	}
	var avatarURL string
	if user.PhotoURL != nil {
		avatarURL = *user.PhotoURL
	}

	// Flatten totals from metrics
	var totalCommits, totalPRs, totalReviews, totalIssues int
	var linesAdded, linesDeleted int64
	if metrics != nil {
		totalCommits = metrics.TotalCommits
		totalPRs = metrics.TotalPRs
		totalReviews = metrics.TotalReviews
		totalIssues = metrics.TotalIssues
		linesAdded = metrics.TotalLinesAdded
		linesDeleted = metrics.TotalLinesDeleted
	}

	// Fetch recent activities (always last 7 days, independent of selected period)
	recentStart := time.Now().AddDate(0, 0, -7)
	recentEnd := time.Now()
	recentActivitiesRaw, err := s.activityRepo.GetByUser(ctx, userID, recentStart, recentEnd, 20, 0)
	if err != nil {
		recentActivitiesRaw = nil
	}

	recentActivities := make([]map[string]interface{}, 0, len(recentActivitiesRaw))
	for _, a := range recentActivitiesRaw {
		title := s.activityTitle(a)
		var url *string
		if a.PRURL != nil {
			url = a.PRURL
		} else if a.IssueURL != nil {
			url = a.IssueURL
		}
		repoFullName := a.RepositoryOwner + "/" + a.RepositoryName
		entry := map[string]interface{}{
			"id":              a.ID,
			"type":            string(a.ActivityType),
			"title":           title,
			"repository_name": repoFullName,
			"created_at":      a.OccurredAt.Format(time.RFC3339),
		}
		if url != nil {
			entry["url"] = *url
		}
		recentActivities = append(recentActivities, entry)
	}

	// Build flat response matching frontend UserAnalytics interface
	result := map[string]interface{}{
		"user_id":    user.ID,
		"user_name":  user.FirstName + " " + user.LastName,
		"user_email": userEmail,
		"avatar_url": avatarURL,

		"total_commits":  totalCommits,
		"total_prs":      totalPRs,
		"total_reviews":  totalReviews,
		"total_issues":   totalIssues,
		"lines_added":    linesAdded,
		"lines_deleted":  linesDeleted,
		"score":          score,

		"teams":                    teamNames,
		"repository_contributions": paginatedRepos,
		"repository_contributions_pagination": map[string]interface{}{
			"total":       totalRepos,
			"page":        repoPage,
			"page_size":   repoPageSize,
			"total_pages": totalRepoPages,
		},
		"activity_timeline":  activityTimeline,
		"recent_activities":  recentActivities,
		"period_start":       startDate.Format(time.RFC3339),
		"period_end":         endDate.Format(time.RFC3339),
	}

	return result, nil
}

// activityTitle returns a human-readable title for a git activity.
func (s *AnalyticsService) activityTitle(a entity.GitActivity) string {
	repo := a.RepositoryOwner + "/" + a.RepositoryName
	switch a.ActivityType {
	case entity.ActivityTypeCommit:
		if a.CommitCount > 1 {
			return fmt.Sprintf("%d commits in %s", a.CommitCount, repo)
		}
		return fmt.Sprintf("Commit in %s", repo)
	case entity.ActivityTypePR:
		if a.PRTitle != nil && *a.PRTitle != "" {
			return *a.PRTitle
		}
		return fmt.Sprintf("Pull request in %s", repo)
	case entity.ActivityTypeReview:
		return fmt.Sprintf("Code review in %s", repo)
	case entity.ActivityTypeIssue:
		if a.IssueTitle != nil && *a.IssueTitle != "" {
			return *a.IssueTitle
		}
		return fmt.Sprintf("Issue in %s", repo)
	default:
		return fmt.Sprintf("Activity in %s", repo)
	}
}

// GetTeamAnalytics retrieves detailed analytics for a team
func (s *AnalyticsService) GetTeamAnalytics(ctx context.Context, teamID uuid.UUID, startDate, endDate time.Time, repoPage, repoPageSize int) (map[string]interface{}, error) {
	// Get team with members
	team, err := s.teamRepo.GetTeamWithMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Get team metrics
	teamMetrics, err := s.metricsRepo.GetByTeamAndPeriod(ctx, teamID, startDate, endDate)
	if err != nil {
		// Try to aggregate on-the-fly
		if err := s.aggService.AggregateTeamMetrics(ctx, teamID, startDate, endDate); err != nil {
			return nil, fmt.Errorf("failed to get team metrics: %w", err)
		}
		teamMetrics, err = s.metricsRepo.GetByTeamAndPeriod(ctx, teamID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get team metrics: %w", err)
		}
	}

	// Get member metrics
	memberMetrics, err := s.metricsRepo.GetTeamMembersMetrics(ctx, teamID, startDate, endDate)
	if err != nil {
		memberMetrics = []entity.AggregatedMetrics{}
	}

	// Calculate team score (may be recalculated below if totals come from member sum)
	var score float64
	if teamMetrics != nil {
		score = s.aggService.CalculateScore(teamMetrics)
	}

	// Get activity timeline data
	activityTimeline, err := s.getActivityTimeline(ctx, teamID, startDate, endDate)
	if err != nil {
		// If we can't get timeline data, continue with empty array
		activityTimeline = []map[string]interface{}{}
	}

	// Get team repositories from DB
	teamRepos, _ := s.getTeamRepoContributions(ctx, teamID, startDate, endDate)

	// Calculate repository pagination
	totalRepos := len(teamRepos)
	totalRepoPages := (totalRepos + repoPageSize - 1) / repoPageSize
	repoOffset := (repoPage - 1) * repoPageSize
	repoEnd := repoOffset + repoPageSize
	if repoEnd > totalRepos {
		repoEnd = totalRepos
	}

	paginatedRepos := teamRepos
	if repoOffset < totalRepos {
		paginatedRepos = teamRepos[repoOffset:repoEnd]
	} else {
		paginatedRepos = []map[string]interface{}{}
	}

	// Flatten totals from team metrics; fall back to summing member metrics
	// when the team-level aggregation row is missing or empty.
	var totalCommits, totalPRs, totalReviews, totalIssues int
	var linesAdded, linesDeleted int64
	if teamMetrics != nil && teamMetrics.TotalCommits > 0 {
		totalCommits = teamMetrics.TotalCommits
		totalPRs = teamMetrics.TotalPRs
		totalReviews = teamMetrics.TotalReviews
		totalIssues = teamMetrics.TotalIssues
		linesAdded = teamMetrics.TotalLinesAdded
		linesDeleted = teamMetrics.TotalLinesDeleted
	} else {
		// Sum from individual member metrics
		for _, mm := range memberMetrics {
			totalCommits += mm.TotalCommits
			totalPRs += mm.TotalPRs
			totalReviews += mm.TotalReviews
			totalIssues += mm.TotalIssues
			linesAdded += mm.TotalLinesAdded
			linesDeleted += mm.TotalLinesDeleted
		}
		// Recalculate score from the summed totals
		syntheticMetrics := &entity.AggregatedMetrics{
			TotalCommits:      totalCommits,
			TotalPRs:          totalPRs,
			TotalReviews:      totalReviews,
			TotalIssues:       totalIssues,
			TotalLinesAdded:   linesAdded,
			TotalLinesDeleted: linesDeleted,
		}
		score = s.aggService.CalculateScore(syntheticMetrics)
	}

	// Build member_stats with user details.
	// When pre-aggregated metrics are available, use them; otherwise fall back
	// to computing stats from raw activities (same approach as the dashboard).
	memberStats := make([]map[string]interface{}, 0)

	if len(memberMetrics) > 0 {
		// Use pre-aggregated metrics.
		var memberUserIDs []uuid.UUID
		for _, mm := range memberMetrics {
			if mm.UserID != nil {
				memberUserIDs = append(memberUserIDs, *mm.UserID)
			}
		}
		memberUsers, _ := s.userRepo.GetByIDs(ctx, memberUserIDs)
		memberUserMap := make(map[uuid.UUID]entity.User)
		for _, u := range memberUsers {
			memberUserMap[u.ID] = u
		}

		for i, mm := range memberMetrics {
			if mm.UserID == nil {
				continue
			}
			mu, ok := memberUserMap[*mm.UserID]
			if !ok {
				continue
			}
			var muAvatar string
			if mu.PhotoURL != nil {
				muAvatar = *mu.PhotoURL
			}
			var muEmail string
			if mu.Email != nil {
				muEmail = *mu.Email
			}
			memberScore := s.aggService.CalculateScore(&mm)
			memberStats = append(memberStats, map[string]interface{}{
				"user_id":       mu.ID,
				"user_name":     mu.FirstName + " " + mu.LastName,
				"user_email":    muEmail,
				"avatar_url":    muAvatar,
				"commits":       mm.TotalCommits,
				"prs":           mm.TotalPRs,
				"reviews":       mm.TotalReviews,
				"issues":        mm.TotalIssues,
				"lines_added":   mm.TotalLinesAdded,
				"lines_deleted": mm.TotalLinesDeleted,
				"score":         memberScore,
				"rank":          i + 1,
			})
		}
	} else if len(team.Members) > 0 {
		// Fallback: compute member stats from raw activities.
		var memberIDs []uuid.UUID
		for _, m := range team.Members {
			memberIDs = append(memberIDs, m.UserID)
		}
		memberActivities, _ := s.activityRepo.GetByPeriod(ctx, startDate, endDate)
		memberIDSet := make(map[uuid.UUID]bool)
		for _, id := range memberIDs {
			memberIDSet[id] = true
		}

		userStats := make(map[uuid.UUID]map[string]int64)
		for _, a := range memberActivities {
			if !memberIDSet[a.UserID] {
				continue
			}
			if _, ok := userStats[a.UserID]; !ok {
				userStats[a.UserID] = map[string]int64{"commits": 0, "prs": 0, "reviews": 0, "issues": 0, "lines_added": 0, "lines_deleted": 0}
			}
			us := userStats[a.UserID]
			switch a.ActivityType {
			case entity.ActivityTypeCommit:
				us["commits"] += int64(a.CommitCount)
				us["lines_added"] += int64(a.LinesAdded)
				us["lines_deleted"] += int64(a.LinesDeleted)
			case entity.ActivityTypePR:
				us["prs"]++
				us["lines_added"] += int64(a.LinesAdded)
				us["lines_deleted"] += int64(a.LinesDeleted)
			case entity.ActivityTypeReview:
				us["reviews"]++
			case entity.ActivityTypeIssue:
				us["issues"]++
			}
		}

		// Build sorted member stats.
		type mScore struct {
			Member entity.TeamMember
			Stats  map[string]int64
			Score  float64
		}
		var mScores []mScore
		for _, m := range team.Members {
			st := userStats[m.UserID]
			if st == nil {
				st = map[string]int64{"commits": 0, "prs": 0, "reviews": 0, "issues": 0, "lines_added": 0, "lines_deleted": 0}
			}
			sc := float64(st["commits"])*1.0 + float64(st["lines_added"])*0.01 + float64(st["prs"])*5.0 + float64(st["reviews"])*3.0 + float64(st["issues"])*2.0
			mScores = append(mScores, mScore{Member: m, Stats: st, Score: sc})
		}
		sort.Slice(mScores, func(i, j int) bool { return mScores[i].Score > mScores[j].Score })

		for rank, ms := range mScores {
			mu := ms.Member.User
			var muName, muAvatar, muEmail string
			if mu != nil {
				muName = mu.FirstName + " " + mu.LastName
				if mu.PhotoURL != nil {
					muAvatar = *mu.PhotoURL
				}
				if mu.Email != nil {
					muEmail = *mu.Email
				}
			}
			memberStats = append(memberStats, map[string]interface{}{
				"user_id":       ms.Member.UserID,
				"user_name":     muName,
				"user_email":    muEmail,
				"avatar_url":    muAvatar,
				"commits":       int(ms.Stats["commits"]),
				"prs":           int(ms.Stats["prs"]),
				"reviews":       int(ms.Stats["reviews"]),
				"issues":        int(ms.Stats["issues"]),
				"lines_added":   ms.Stats["lines_added"],
				"lines_deleted": ms.Stats["lines_deleted"],
				"score":         ms.Score,
				"rank":          rank + 1,
			})
		}
	}

	// Resolve lead info
	var leadID *uuid.UUID
	var leadName string
	if team.ManagerID != nil {
		leadID = team.ManagerID
		if manager, err := s.userRepo.GetByID(ctx, *team.ManagerID); err == nil {
			leadName = manager.FirstName + " " + manager.LastName
		}
	}

	// Build flat response matching frontend TeamAnalytics interface
	result := map[string]interface{}{
		"team_id":   team.ID,
		"team_name": team.Name,
		"lead_id":   leadID,
		"lead_name": leadName,

		"total_commits":  totalCommits,
		"total_prs":      totalPRs,
		"total_reviews":  totalReviews,
		"total_issues":   totalIssues,
		"lines_added":    linesAdded,
		"lines_deleted":  linesDeleted,
		"score":          score,

		"member_stats":     memberStats,
		"repository_stats": paginatedRepos,
		"repository_stats_pagination": map[string]interface{}{
			"total":       totalRepos,
			"page":        repoPage,
			"page_size":   repoPageSize,
			"total_pages": totalRepoPages,
		},
		"activity_timeline": activityTimeline,
		"period_start":      startDate.Format(time.RFC3339),
		"period_end":        endDate.Format(time.RFC3339),
	}

	return result, nil
}

// GetRepositoryAnalytics retrieves analytics for a specific repository
func (s *AnalyticsService) GetRepositoryAnalytics(ctx context.Context, owner, name string, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get activities for the repository
	activities, err := s.activityRepo.GetByRepository(ctx, owner, name, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository activities: %w", err)
	}

	// Calculate statistics
	stats := make(map[string]interface{})
	contributorsMap := make(map[uuid.UUID]map[string]int)

	for _, activity := range activities {
		// Track contributor stats
		if _, exists := contributorsMap[activity.UserID]; !exists {
			contributorsMap[activity.UserID] = map[string]int{
				"commits":       0,
				"lines_added":   0,
				"lines_deleted": 0,
				"prs":           0,
				"reviews":       0,
			}
		}

		contributor := contributorsMap[activity.UserID]
		switch activity.ActivityType {
		case entity.ActivityTypeCommit:
			contributor["commits"] += activity.CommitCount
			contributor["lines_added"] += activity.LinesAdded
			contributor["lines_deleted"] += activity.LinesDeleted
		case entity.ActivityTypePR:
			contributor["prs"]++
		case entity.ActivityTypeReview:
			contributor["reviews"]++
		}
	}

	// Get contributor details
	var contributorIDs []uuid.UUID
	for userID := range contributorsMap {
		contributorIDs = append(contributorIDs, userID)
	}

	users, _ := s.userRepo.GetByIDs(ctx, contributorIDs)

	// Build contributors list
	var contributors []map[string]interface{}
	for _, user := range users {
		if contribStats, ok := contributorsMap[user.ID]; ok {
			contributors = append(contributors, map[string]interface{}{
				"user":  user,
				"stats": contribStats,
			})
		}
	}

	stats["repository"] = map[string]string{
		"owner": owner,
		"name":  name,
	}
	stats["contributors"] = contributors
	stats["total_activities"] = len(activities)
	stats["period"] = map[string]time.Time{
		"start": startDate,
		"end":   endDate,
	}

	return stats, nil
}

// GetDashboardStats retrieves summary statistics for the dashboard.
// Top contributors and teams are computed directly from raw activities so the
// dashboard works even before/without aggregation.
func (s *AnalyticsService) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get periods summary
	periodsSummary, _ := s.metricsRepo.GetPeriodsSummary(ctx)

	// Count active users
	activeFlag := true
	activeUsers, _ := s.userRepo.Count(ctx, &activeFlag)

	// Count active teams
	teams, _ := s.teamRepo.List(ctx, &activeFlag, 0, 0)

	// Get dashboard activity trend (from raw activities).
	activityTrend, err := s.getDashboardActivityTrend(ctx, startDate, endDate)
	if err != nil {
		activityTrend = []map[string]interface{}{}
	}

	// Calculate totals from the activity trend data.
	var totalCommits, totalPRs, totalReviews, totalIssues int
	for _, day := range activityTrend {
		if v, ok := day["commits"].(int); ok {
			totalCommits += v
		}
		if v, ok := day["prs"].(int); ok {
			totalPRs += v
		}
		if v, ok := day["reviews"].(int); ok {
			totalReviews += v
		}
		if v, ok := day["issues"].(int); ok {
			totalIssues += v
		}
	}

	// ----- Top Contributors: computed from raw activities -----
	activities, _ := s.activityRepo.GetByPeriod(ctx, startDate, endDate)

	userStats := make(map[uuid.UUID]map[string]int)
	for _, a := range activities {
		if _, ok := userStats[a.UserID]; !ok {
			userStats[a.UserID] = map[string]int{"commits": 0, "prs": 0, "reviews": 0, "issues": 0, "lines_added": 0, "lines_deleted": 0}
		}
		us := userStats[a.UserID]
		switch a.ActivityType {
		case entity.ActivityTypeCommit:
			us["commits"] += a.CommitCount
			us["lines_added"] += a.LinesAdded
			us["lines_deleted"] += a.LinesDeleted
		case entity.ActivityTypePR:
			us["prs"]++
		case entity.ActivityTypeReview:
			us["reviews"]++
		case entity.ActivityTypeIssue:
			us["issues"]++
		}
	}

	// Sort users by score, take top 10.
	type userScore struct {
		UserID uuid.UUID
		Stats  map[string]int
		Score  float64
	}
	var userScores []userScore
	for uid, st := range userStats {
		score := float64(st["commits"])*1.0 +
			float64(st["lines_added"])*0.01 +
			float64(st["prs"])*5.0 +
			float64(st["reviews"])*3.0 +
			float64(st["issues"])*2.0
		userScores = append(userScores, userScore{UserID: uid, Stats: st, Score: score})
	}
	sort.Slice(userScores, func(i, j int) bool { return userScores[i].Score > userScores[j].Score })

	// Fetch user details.
	var topUserIDs []uuid.UUID
	for _, us := range userScores {
		topUserIDs = append(topUserIDs, us.UserID)
	}
	topUserDetails, _ := s.userRepo.GetByIDs(ctx, topUserIDs)
	topUserMap := make(map[uuid.UUID]entity.User)
	for _, u := range topUserDetails {
		topUserMap[u.ID] = u
	}

	topContributors := make([]map[string]interface{}, 0, len(userScores))
	for rank, us := range userScores {
		user, ok := topUserMap[us.UserID]
		if !ok {
			continue
		}
		var avatarURL string
		if user.PhotoURL != nil {
			avatarURL = *user.PhotoURL
		}
		topContributors = append(topContributors, map[string]interface{}{
			"user_id":       user.ID,
			"user_name":     user.FirstName + " " + user.LastName,
			"avatar_url":    avatarURL,
			"commits":       us.Stats["commits"],
			"prs":           us.Stats["prs"],
			"reviews":       us.Stats["reviews"],
			"issues":        us.Stats["issues"],
			"lines_added":   us.Stats["lines_added"],
			"lines_deleted": us.Stats["lines_deleted"],
			"rank":          rank + 1,
		})
	}

	// ----- Top Teams: sum member activities -----
	topTeamsList := make([]map[string]interface{}, 0)
	if len(teams) > 0 {
		type teamScore struct {
			Team    entity.Team
			Stats   map[string]int
			Members int
			Score   float64
		}
		var teamScores []teamScore
		for _, team := range teams {
			members, _ := s.teamRepo.GetMembers(ctx, team.ID)
			st := map[string]int{"commits": 0, "prs": 0, "reviews": 0, "issues": 0, "lines_added": 0, "lines_deleted": 0}
			for _, m := range members {
				if us, ok := userStats[m.UserID]; ok {
					for k, v := range us {
						st[k] += v
					}
				}
			}
			score := float64(st["commits"])*1.0 + float64(st["lines_added"])*0.01 + float64(st["prs"])*5.0 + float64(st["reviews"])*3.0 + float64(st["issues"])*2.0
			teamScores = append(teamScores, teamScore{Team: team, Stats: st, Members: len(members), Score: score})
		}
		sort.Slice(teamScores, func(i, j int) bool { return teamScores[i].Score > teamScores[j].Score })

		for rank, ts := range teamScores {
			topTeamsList = append(topTeamsList, map[string]interface{}{
				"team_id":       ts.Team.ID,
				"team_name":     ts.Team.Name,
				"total_commits": ts.Stats["commits"],
				"total_prs":     ts.Stats["prs"],
				"total_reviews": ts.Stats["reviews"],
				"total_issues":  ts.Stats["issues"],
				"lines_added":   ts.Stats["lines_added"],
				"lines_deleted": ts.Stats["lines_deleted"],
				"member_count":  ts.Members,
				"rank":          rank + 1,
			})
		}
	}

	dashboard := map[string]interface{}{
		"top_contributors": topContributors,
		"top_teams":        topTeamsList,
		"total_users":      activeUsers,
		"total_teams":      len(teams),
		"total_commits":    totalCommits,
		"total_prs":        totalPRs,
		"total_reviews":    totalReviews,
		"total_issues":     totalIssues,
		"periods":          periodsSummary,
		"activity_trend":   activityTrend,
		"period_start":     startDate.Format(time.RFC3339),
		"period_end":       endDate.Format(time.RFC3339),
	}

	return dashboard, nil
}

// GetRepositoriesLeaderboard retrieves repositories leaderboard with activity timeline
func (s *AnalyticsService) GetRepositoriesLeaderboard(ctx context.Context, startDate, endDate time.Time, page, pageSize int) (map[string]interface{}, error) {
	// Get all activities within date range to calculate repository stats
	activities, err := s.activityRepo.GetByPeriod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository activities: %w", err)
	}

	// Aggregate activities by repository
	repoStats := make(map[string]map[string]int)
	contributorCounts := make(map[string]map[string]bool) // repo -> user_id -> true

	for _, activity := range activities {
		repoKey := fmt.Sprintf("%s/%s", activity.RepositoryOwner, activity.RepositoryName)
		if _, exists := repoStats[repoKey]; !exists {
			repoStats[repoKey] = map[string]int{
				"commits":       0,
				"prs":           0,
				"reviews":       0,
				"issues":        0,
				"lines_added":   0,
				"lines_deleted": 0,
			}
			contributorCounts[repoKey] = make(map[string]bool)
		}

		stats := repoStats[repoKey]
		contributorCounts[repoKey][activity.UserID.String()] = true

		switch activity.ActivityType {
		case entity.ActivityTypeCommit:
			stats["commits"] += activity.CommitCount
			stats["lines_added"] += activity.LinesAdded
			stats["lines_deleted"] += activity.LinesDeleted
		case entity.ActivityTypePR:
			stats["prs"]++
			stats["lines_added"] += activity.LinesAdded
			stats["lines_deleted"] += activity.LinesDeleted
		case entity.ActivityTypeReview:
			stats["reviews"]++
		case entity.ActivityTypeIssue:
			stats["issues"]++
		}
	}

	// Convert to slice and calculate activity score
	type repositoryStats struct {
		Name          string
		TotalCommits  int
		TotalPRs      int
		Contributors  int
		LinesAdded    int
		LinesDeleted  int
		ActivityScore float64
		Rank          int
	}

	repositories := make([]repositoryStats, 0, len(repoStats))
	for repoKey, stats := range repoStats {
		// Calculate activity score (weighted formula)
		activityScore := float64(stats["commits"])*1.0 +
			float64(stats["prs"])*3.0 +
			float64(stats["reviews"])*1.5 +
			float64(stats["issues"])*1.0 +
			float64(len(contributorCounts[repoKey]))*2.0

		repositories = append(repositories, repositoryStats{
			Name:          repoKey,
			TotalCommits:  stats["commits"],
			TotalPRs:      stats["prs"],
			Contributors:  len(contributorCounts[repoKey]),
			LinesAdded:    stats["lines_added"],
			LinesDeleted:  stats["lines_deleted"],
			ActivityScore: activityScore,
		})
	}

	// Sort by activity score descending, then by name for deterministic order
	sort.Slice(repositories, func(i, j int) bool {
		if repositories[i].ActivityScore != repositories[j].ActivityScore {
			return repositories[i].ActivityScore > repositories[j].ActivityScore
		}
		return repositories[i].Name < repositories[j].Name
	})

	// Set ranks
	for i := range repositories {
		repositories[i].Rank = i + 1
	}

	// Calculate pagination
	totalCount := len(repositories)
	totalPages := (totalCount + pageSize - 1) / pageSize
	offset := (page - 1) * pageSize

	// Apply pagination
	paginatedRepos := repositories
	if offset < totalCount {
		end := offset + pageSize
		if end > totalCount {
			end = totalCount
		}
		paginatedRepos = repositories[offset:end]
	} else {
		paginatedRepos = []repositoryStats{}
	}

	// Convert to response format
	repoList := make([]map[string]interface{}, len(paginatedRepos))
	for i, repo := range paginatedRepos {
		repoList[i] = map[string]interface{}{
			"repository_name":    repo.Name,
			"total_commits":      repo.TotalCommits,
			"total_prs":          repo.TotalPRs,
			"contributors_count": repo.Contributors,
			"lines_added":        repo.LinesAdded,
			"lines_deleted":      repo.LinesDeleted,
			"activity_score":     repo.ActivityScore,
			"rank":               repo.Rank,
		}
	}

	// Get repository activity timeline (sum of all repositories by day)
	activityTimeline, err := s.getRepositoryActivityTimeline(ctx, startDate, endDate)
	if err != nil {
		// If we can't get timeline data, continue with empty array
		activityTimeline = []map[string]interface{}{}
	}

	// Compute global totals from ALL repos (not just the current page).
	var globalCommits, globalPRs, globalLinesAdded, globalLinesDeleted int
	globalContributors := make(map[string]bool)
	for repoKey, stats := range repoStats {
		globalCommits += stats["commits"]
		globalPRs += stats["prs"]
		globalLinesAdded += stats["lines_added"]
		globalLinesDeleted += stats["lines_deleted"]
		for uid := range contributorCounts[repoKey] {
			globalContributors[uid] = true
		}
	}

	// Build response
	result := map[string]interface{}{
		"repositories":                 repoList,
		"total":                        totalCount,
		"page":                         page,
		"page_size":                    pageSize,
		"total_pages":                  totalPages,
		"total_commits":                globalCommits,
		"total_prs":                    globalPRs,
		"total_lines_added":            globalLinesAdded,
		"total_lines_deleted":          globalLinesDeleted,
		"total_contributors":           len(globalContributors),
		"period_start":                 startDate.Format(time.RFC3339),
		"period_end":                   endDate.Format(time.RFC3339),
		"repository_activity_timeline": activityTimeline,
	}

	return result, nil
}

// getActivityTimeline generates daily activity timeline data
func (s *AnalyticsService) getActivityTimeline(ctx context.Context, teamID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get team members
	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	if len(members) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Collect member user IDs
	var userIDs []uuid.UUID
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	// Get activities for team members within date range
	activities, err := s.activityRepo.GetByUsersAndPeriod(ctx, userIDs, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	// Create a map to aggregate daily activities
	dailyActivity := make(map[string]map[string]int)

	// Initialize map with all dates in range
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		dailyActivity[dateStr] = map[string]int{
			"commits": 0,
			"prs":     0,
			"reviews": 0,
			"issues":  0,
		}
		current = current.AddDate(0, 0, 1)
	}

	// Aggregate activities by date
	for _, activity := range activities {
		dateStr := activity.OccurredAt.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			switch activity.ActivityType {
			case entity.ActivityTypeCommit:
				dayData["commits"] += activity.CommitCount
			case entity.ActivityTypePR:
				dayData["prs"]++
			case entity.ActivityTypeReview:
				dayData["reviews"]++
			case entity.ActivityTypeIssue:
				dayData["issues"]++
			}
		}
	}

	// Convert to sorted slice
	var timeline []map[string]interface{}
	current = startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			timeline = append(timeline, map[string]interface{}{
				"date":    dateStr,
				"commits": dayData["commits"],
				"prs":     dayData["prs"],
				"reviews": dayData["reviews"],
				"issues":  dayData["issues"],
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	return timeline, nil
}

// getUserActivityTimeline generates daily activity timeline data for a user
func (s *AnalyticsService) getUserActivityTimeline(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get activities for the user within date range
	activities, err := s.activityRepo.GetByUser(ctx, userID, startDate, endDate, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user activities: %w", err)
	}

	if len(activities) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Create a map to aggregate daily activities
	dailyActivity := make(map[string]map[string]int)

	// Initialize map with all dates in range
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		dailyActivity[dateStr] = map[string]int{
			"commits": 0,
			"prs":     0,
			"reviews": 0,
			"issues":  0,
		}
		current = current.AddDate(0, 0, 1)
	}

	// Aggregate activities by date
	for _, activity := range activities {
		dateStr := activity.OccurredAt.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			switch activity.ActivityType {
			case entity.ActivityTypeCommit:
				dayData["commits"] += activity.CommitCount
			case entity.ActivityTypePR:
				dayData["prs"]++
			case entity.ActivityTypeReview:
				dayData["reviews"]++
			case entity.ActivityTypeIssue:
				dayData["issues"]++
			}
		}
	}

	// Convert to sorted slice
	var timeline []map[string]interface{}
	current = startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			timeline = append(timeline, map[string]interface{}{
				"date":    dateStr,
				"commits": dayData["commits"],
				"prs":     dayData["prs"],
				"reviews": dayData["reviews"],
				"issues":  dayData["issues"],
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	return timeline, nil
}

// getRepositoryActivityTimeline generates daily activity timeline data for all repositories combined
func (s *AnalyticsService) getRepositoryActivityTimeline(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get all activities within date range
	activities, err := s.activityRepo.GetByPeriod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository activities: %w", err)
	}

	if len(activities) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Create a map to aggregate daily activities across all repositories
	dailyActivity := make(map[string]map[string]int)

	// Initialize map with all dates in range
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		dailyActivity[dateStr] = map[string]int{
			"commits": 0,
			"prs":     0,
			"reviews": 0,
			"issues":  0,
		}
		current = current.AddDate(0, 0, 1)
	}

	// Aggregate activities by date across all repositories
	for _, activity := range activities {
		dateStr := activity.OccurredAt.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			switch activity.ActivityType {
			case entity.ActivityTypeCommit:
				dayData["commits"] += activity.CommitCount
			case entity.ActivityTypePR:
				dayData["prs"]++
			case entity.ActivityTypeReview:
				dayData["reviews"]++
			case entity.ActivityTypeIssue:
				dayData["issues"]++
			}
		}
	}

	// Convert to sorted slice
	var timeline []map[string]interface{}
	current = startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			timeline = append(timeline, map[string]interface{}{
				"date":    dateStr,
				"commits": dayData["commits"],
				"prs":     dayData["prs"],
				"reviews": dayData["reviews"],
				"issues":  dayData["issues"],
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	return timeline, nil
}

// getUserRepoContributions retrieves repository contribution stats for a user from DB.
func (s *AnalyticsService) getUserRepoContributions(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	activities, err := s.activityRepo.GetByUser(ctx, userID, startDate, endDate, 10000, 0)
	if err != nil {
		return nil, err
	}

	repoMap := make(map[string]map[string]int)
	for _, a := range activities {
		key := a.RepositoryOwner + "/" + a.RepositoryName
		if _, ok := repoMap[key]; !ok {
			repoMap[key] = map[string]int{"commits": 0, "prs": 0, "lines_added": 0, "lines_deleted": 0}
		}
		switch a.ActivityType {
		case entity.ActivityTypeCommit:
			repoMap[key]["commits"] += a.CommitCount
			repoMap[key]["lines_added"] += a.LinesAdded
			repoMap[key]["lines_deleted"] += a.LinesDeleted
		case entity.ActivityTypePR:
			repoMap[key]["prs"]++
		}
	}

	result := make([]map[string]interface{}, 0, len(repoMap))
	for name, stats := range repoMap {
		result = append(result, map[string]interface{}{
			"repository_name": name,
			"commits":         stats["commits"],
			"prs":             stats["prs"],
			"lines_added":     stats["lines_added"],
			"lines_deleted":   stats["lines_deleted"],
		})
	}

	// Sort by commits descending
	sort.Slice(result, func(i, j int) bool {
		return result[i]["commits"].(int) > result[j]["commits"].(int)
	})

	return result, nil
}

// getTeamRepoContributions retrieves repository contribution stats for a team from DB.
func (s *AnalyticsService) getTeamRepoContributions(ctx context.Context, teamID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	activities, err := s.activityRepo.GetByTeam(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	repoMap := make(map[string]map[string]int)
	contributorSets := make(map[string]map[string]bool)
	for _, a := range activities {
		key := a.RepositoryOwner + "/" + a.RepositoryName
		if _, ok := repoMap[key]; !ok {
			repoMap[key] = map[string]int{"commits": 0, "prs": 0, "lines_added": 0, "lines_deleted": 0}
			contributorSets[key] = make(map[string]bool)
		}
		contributorSets[key][a.UserID.String()] = true
		switch a.ActivityType {
		case entity.ActivityTypeCommit:
			repoMap[key]["commits"] += a.CommitCount
			repoMap[key]["lines_added"] += a.LinesAdded
			repoMap[key]["lines_deleted"] += a.LinesDeleted
		case entity.ActivityTypePR:
			repoMap[key]["prs"]++
		}
	}

	result := make([]map[string]interface{}, 0, len(repoMap))
	for name, stats := range repoMap {
		result = append(result, map[string]interface{}{
			"repository_name": name,
			"commits":         stats["commits"],
			"prs":             stats["prs"],
			"contributors":    len(contributorSets[name]),
			"lines_added":     stats["lines_added"],
			"lines_deleted":   stats["lines_deleted"],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["commits"].(int) > result[j]["commits"].(int)
	})

	return result, nil
}

// getDashboardActivityTrend generates global activity trend data
func (s *AnalyticsService) getDashboardActivityTrend(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get all activities within date range
	activities, err := s.activityRepo.GetByPeriod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard activities: %w", err)
	}

	// Create a map to aggregate daily activities
	dailyActivity := make(map[string]map[string]int)

	// Initialize map with all dates in range
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		dailyActivity[dateStr] = map[string]int{
			"commits": 0,
			"prs":     0,
			"reviews": 0,
			"issues":  0,
		}
		current = current.AddDate(0, 0, 1)
	}

	// Aggregate activities by date
	for _, activity := range activities {
		dateStr := activity.OccurredAt.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			switch activity.ActivityType {
			case entity.ActivityTypeCommit:
				dayData["commits"] += activity.CommitCount
			case entity.ActivityTypePR:
				dayData["prs"]++
			case entity.ActivityTypeReview:
				dayData["reviews"]++
			case entity.ActivityTypeIssue:
				dayData["issues"]++
			}
		}
	}

	// Convert to sorted slice
	var timeline []map[string]interface{}
	current = startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayData, exists := dailyActivity[dateStr]; exists {
			timeline = append(timeline, map[string]interface{}{
				"date":    dateStr,
				"commits": dayData["commits"],
				"prs":     dayData["prs"],
				"reviews": dayData["reviews"],
				"issues":  dayData["issues"],
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	return timeline, nil
}
