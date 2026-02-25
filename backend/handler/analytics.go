package handler

import (
	"net/http"
	"strconv"
	"time"

	"gconsus/lib/http/rest"
	"gconsus/service"

	"github.com/google/uuid"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetLeaderboard handles GET /api/v1/analytics/leaderboard
func (h *AnalyticsHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	leaderboard, err := h.analyticsService.GetLeaderboard(r.Context(), startDate, endDate, limit)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"leaderboard": leaderboard,
		"period": map[string]string{
			"start": startDate.Format(time.RFC3339),
			"end":   endDate.Format(time.RFC3339),
		},
		"count": len(leaderboard),
	})
}

// GetTeamLeaderboard handles GET /api/v1/analytics/teams/leaderboard
func (h *AnalyticsHandler) GetTeamLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	leaderboard, err := h.analyticsService.GetTeamLeaderboard(r.Context(), startDate, endDate, limit)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"leaderboard": leaderboard,
		"period": map[string]string{
			"start": startDate.Format(time.RFC3339),
			"end":   endDate.Format(time.RFC3339),
		},
		"count": len(leaderboard),
	})
}

// GetUserAnalytics handles GET /api/v1/analytics/users/{id}
func (h *AnalyticsHandler) GetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	repoPage, _ := strconv.Atoi(r.URL.Query().Get("repo_page"))
	if repoPage <= 0 {
		repoPage = 1
	}

	repoPageSize, _ := strconv.Atoi(r.URL.Query().Get("repo_page_size"))
	if repoPageSize <= 0 || repoPageSize > 100 {
		repoPageSize = 25
	}

	analytics, err := h.analyticsService.GetUserAnalytics(r.Context(), userID, startDate, endDate, repoPage, repoPageSize)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, analytics)
}

// GetTeamAnalytics handles GET /api/v1/analytics/teams/{id}
func (h *AnalyticsHandler) GetTeamAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	repoPage, _ := strconv.Atoi(r.URL.Query().Get("repo_page"))
	if repoPage <= 0 {
		repoPage = 1
	}

	repoPageSize, _ := strconv.Atoi(r.URL.Query().Get("repo_page_size"))
	if repoPageSize <= 0 || repoPageSize > 100 {
		repoPageSize = 25
	}

	analytics, err := h.analyticsService.GetTeamAnalytics(r.Context(), teamID, startDate, endDate, repoPage, repoPageSize)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, analytics)
}

// GetRepositoryAnalytics handles GET /api/v1/analytics/repositories
func (h *AnalyticsHandler) GetRepositoryAnalytics(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	name := r.URL.Query().Get("name")

	if owner == "" || name == "" {
		rest.ReturnRequestError(w, "owner and name query parameters are required")
		return
	}

	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	analytics, err := h.analyticsService.GetRepositoryAnalytics(r.Context(), owner, name, startDate, endDate)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, analytics)
}

// GetRepositoriesLeaderboard handles GET /api/v1/analytics/repositories/leaderboard
func (h *AnalyticsHandler) GetRepositoriesLeaderboard(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	leaderboard, err := h.analyticsService.GetRepositoriesLeaderboard(r.Context(), startDate, endDate, page, pageSize)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, leaderboard)
}

// GetDashboard handles GET /api/v1/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}

	dashboard, err := h.analyticsService.GetDashboardStats(r.Context(), startDate, endDate)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, dashboard)
}

// parseDateRange parses start and end date from query parameters
// Defaults to last 30 days if not specified
func parseDateRange(r *http.Request) (start, end time.Time, err error) {
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		startStr = r.URL.Query().Get("start_date") // frontend compat
	}
	endStr := r.URL.Query().Get("end")
	if endStr == "" {
		endStr = r.URL.Query().Get("end_date") // frontend compat
	}

	now := time.Now().UTC()

	// Default to last 30 days (truncated to day boundaries).
	if startStr == "" {
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -30)
	} else {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	}

	if endStr == "" {
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	} else {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
	}

	// Validate date range
	if start.After(end) {
		return time.Time{}, time.Time{}, err
	}

	return start, end, nil
}
