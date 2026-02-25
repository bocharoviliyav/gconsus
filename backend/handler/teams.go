package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"gconsus/entity"
	"gconsus/lib/http/rest"
	"gconsus/service"

	"github.com/google/uuid"
)

// TeamsHandler handles team-related HTTP requests
type TeamsHandler struct {
	teamService *service.TeamService
}

// NewTeamsHandler creates a new teams handler
func NewTeamsHandler(teamService *service.TeamService) *TeamsHandler {
	return &TeamsHandler{
		teamService: teamService,
	}
}

// CreateTeam handles POST /api/v1/teams
func (h *TeamsHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req entity.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	team, err := h.teamService.CreateTeam(r.Context(), req)
	if err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		slog.Error("Failed to create team", "error", err)
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnCreateResponse(w, teamToResponse(*team, nil))
}

// GetTeam handles GET /api/v1/teams/{id}
func (h *TeamsHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	teamWithMembers, err := h.teamService.GetTeam(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get team", "error", err, "id", id)
		rest.ReturnServerError(w)
		return
	}

	info := h.teamService.GetTeamEnrichedInfo(r.Context(), teamWithMembers.Team)
	rest.ReturnResponse(w, teamToResponse(teamWithMembers.Team, &info))
}

// ListTeams handles GET /api/v1/teams
func (h *TeamsHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var isActive *bool
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			isActive = &active
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Also support legacy limit/offset
	if r.URL.Query().Get("limit") != "" {
		pageSize, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}
	}

	offset := (page - 1) * pageSize

	teams, err := h.teamService.ListTeams(r.Context(), isActive, pageSize, offset)
	if err != nil {
		slog.Error("Failed to list teams", "error", err)
		rest.ReturnServerError(w)
		return
	}

	// Count total for pagination.
	total := len(teams)
	if total == pageSize {
		// There may be more — get full count.
		allTeams, _ := h.teamService.ListTeams(r.Context(), isActive, 0, 0)
		total = len(allTeams)
	} else {
		total = offset + len(teams)
	}

	// Transform to snake_case format expected by frontend.
	teamList := make([]map[string]interface{}, len(teams))
	for i, t := range teams {
		info := h.teamService.GetTeamEnrichedInfo(r.Context(), t)
		teamList[i] = teamToResponse(t, &info)
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"teams":     teamList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateTeam handles PUT /api/v1/teams/{id}
func (h *TeamsHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	var req entity.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	team, err := h.teamService.UpdateTeam(r.Context(), id, req)
	if err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		slog.Error("Failed to update team", "error", err, "id", id)
		rest.ReturnServerError(w)
		return
	}

	info := h.teamService.GetTeamEnrichedInfo(r.Context(), *team)
	rest.ReturnResponse(w, teamToResponse(*team, &info))
}

// DeleteTeam handles DELETE /api/v1/teams/{id}
func (h *TeamsHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	if err := h.teamService.DeleteTeam(r.Context(), id); err != nil {
		slog.Error("Failed to delete team", "error", err, "id", id)
		rest.ReturnServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddTeamMember handles POST /api/v1/teams/{id}/members
func (h *TeamsHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	var req entity.AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	if err := h.teamService.AddTeamMember(r.Context(), teamID, req); err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		slog.Error("Failed to add team member", "error", err, "team_id", teamID)
		rest.ReturnServerError(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	rest.ReturnResponse(w, map[string]string{"message": "member added successfully"})
}

// RemoveTeamMember handles DELETE /api/v1/teams/{id}/members/{userId}
func (h *TeamsHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	teamIDStr := r.PathValue("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	if err := h.teamService.RemoveTeamMember(r.Context(), teamID, userID); err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		rest.ReturnServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTeamMembers handles GET /api/v1/teams/{id}/members
func (h *TeamsHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	members, err := h.teamService.GetTeamMembers(r.Context(), teamID)
	if err != nil {
		slog.Error("Failed to get team members", "error", err, "team_id", teamID)
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"members": members,
		"count":   len(members),
	})
}

// teamToResponse converts entity.Team to the snake_case format the frontend expects.
func teamToResponse(t entity.Team, info *service.TeamEnrichedInfo) map[string]interface{} {
	memberCount := 0
	leadName := ""
	if info != nil {
		memberCount = info.MemberCount
		leadName = info.LeadName
	}
	resp := map[string]interface{}{
		"id":           t.ID,
		"name":         t.Name,
		"description":  t.Description,
		"is_active":    t.IsActive,
		"member_count": memberCount,
		"lead_name":    leadName,
		"created_at":   t.CreatedAt.Format(time.RFC3339),
		"updated_at":   t.UpdatedAt.Format(time.RFC3339),
	}
	if t.ManagerID != nil {
		resp["lead_id"] = t.ManagerID
	}
	return resp
}

// UpdateMemberRole handles PATCH /api/v1/teams/{id}/members/{userId}
func (h *TeamsHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	teamIDStr := r.PathValue("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid team ID")
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	if err := h.teamService.UpdateMemberRole(r.Context(), teamID, userID, req.Role); err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]string{"message": "role updated successfully"})
}
