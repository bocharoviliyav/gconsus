package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gconsus/entity"
	"gconsus/lib/http/rest"
	"gconsus/service"

	"github.com/google/uuid"
)

// UsersHandler handles user-related HTTP requests
type UsersHandler struct {
	userService *service.UserService
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(userService *service.UserService) *UsersHandler {
	return &UsersHandler{
		userService: userService,
	}
}

// CreateUser handles POST /api/v1/users
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req entity.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnCreateResponse(w, user)
}

// GetUser handles GET /api/v1/users/{id}
func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, user)
}

// ListUsers handles GET /api/v1/users
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	var isActive *bool
	if activeStr := query.Get("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			isActive = &active
		}
	}

	search := query.Get("search")

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	users, err := h.userService.ListUsers(r.Context(), search, isActive, limit, offset)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	var req entity.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, req)
	if err != nil {
		var pErr service.ParamsError
		if errors.As(err, &pErr) {
			rest.ReturnRequestError(w, pErr.Message)
			return
		}
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, user)
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		rest.ReturnServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SyncUsers handles POST /api/v1/users/sync
// Triggers synchronization with external employee API
func (h *UsersHandler) SyncUsers(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userService.SyncUsersFromExternalAPI(r.Context())
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"message":        "sync completed",
		"users_created":  stats.Created,
		"users_updated":  stats.Updated,
		"users_disabled": stats.Disabled,
		"total":          stats.Total,
	})
}

// GetUserByUsername handles GET /api/v1/users/username/{username}
func (h *UsersHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		rest.ReturnRequestError(w, "username parameter required")
		return
	}

	user, err := h.userService.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			rest.ReturnNotFound(w, "user not found")
			return
		}
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, user)
}

// GetUserTeams handles GET /api/v1/users/{id}/teams
func (h *UsersHandler) GetUserTeams(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		rest.ReturnRequestError(w, "invalid user ID")
		return
	}

	teams, err := h.userService.GetUserTeams(r.Context(), userID)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"user_id": userID,
		"teams":   teams,
		"count":   len(teams),
	})
}
