package handler

import (
	"encoding/json"
	"net/http"

	"gconsus/entity"
	"gconsus/lib/http/rest"
	"gconsus/service"
)

// SettingsHandler handles settings-related HTTP requests.
type SettingsHandler struct {
	settingsService *service.SettingsService
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(settingsService *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// GetSystemSettings handles GET /settings/system
func (h *SettingsHandler) GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	info, err := h.settingsService.GetSystemInfo(r.Context())
	if err != nil {
		rest.ReturnServerError(w)
		return
	}
	rest.ReturnResponse(w, info)
}

// updateSettingsRequest matches the frontend UpdateSystemSettingsRequest.
type updateSettingsRequest struct {
	GitHubEnabled *bool   `json:"github_enabled"`
	GitHubURL     *string `json:"github_url"`
	GitHubToken   *string `json:"github_token"`
	GitLabEnabled *bool   `json:"gitlab_enabled"`
	GitLabURL     *string `json:"gitlab_url"`
	GitLabToken   *string `json:"gitlab_token"`
	SyncSchedule  *string `json:"sync_schedule_cron"`
	SyncEnabled   *bool   `json:"sync_enabled"`
}

// UpdateSystemSettings handles PUT /settings/system
func (h *SettingsHandler) UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	var req updateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	ctx := r.Context()

	// Upsert GitHub provider if any github_* field is set.
	if req.GitHubEnabled != nil || req.GitHubURL != nil || req.GitHubToken != nil {
		enabled := req.GitHubEnabled != nil && *req.GitHubEnabled
		url := ""
		if req.GitHubURL != nil {
			url = *req.GitHubURL
		}
		token := ""
		if req.GitHubToken != nil {
			token = *req.GitHubToken
		}
		if err := h.settingsService.UpsertProviderSettings(ctx, "github", enabled, url, token); err != nil {
			rest.ReturnServerError(w)
			return
		}
	}

	// Upsert GitLab provider if any gitlab_* field is set.
	if req.GitLabEnabled != nil || req.GitLabURL != nil || req.GitLabToken != nil {
		enabled := req.GitLabEnabled != nil && *req.GitLabEnabled
		url := ""
		if req.GitLabURL != nil {
			url = *req.GitLabURL
		}
		token := ""
		if req.GitLabToken != nil {
			token = *req.GitLabToken
		}
		if err := h.settingsService.UpsertProviderSettings(ctx, "gitlab", enabled, url, token); err != nil {
			rest.ReturnServerError(w)
			return
		}
	}

	// Save schedule configs.
	if req.SyncSchedule != nil {
		valJSON, _ := json.Marshal(*req.SyncSchedule)
		_ = h.settingsService.SetConfig(ctx, "sync_schedule", entity.JSONValue(valJSON), nil)
	}
	if req.SyncEnabled != nil {
		valJSON, _ := json.Marshal(*req.SyncEnabled)
		_ = h.settingsService.SetConfig(ctx, "sync_enabled", entity.JSONValue(valJSON), nil)
	}

	info, err := h.settingsService.GetSystemInfo(ctx)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}
	rest.ReturnResponse(w, info)
}

// ListProviders handles GET /settings/providers
func (h *SettingsHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.settingsService.ListProviders(r.Context(), false)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}
	rest.ReturnResponse(w, map[string]interface{}{"providers": providers})
}

// CreateProvider handles POST /settings/providers
func (h *SettingsHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var p entity.VCSProvider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	if err := h.settingsService.CreateProvider(r.Context(), &p); err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}
	rest.ReturnResponse(w, p)
}

// TestGitHub handles POST /settings/test-github
func (h *SettingsHandler) TestGitHub(w http.ResponseWriter, r *http.Request) {
	h.testConnection(w, r, "github")
}

// TestGitLab handles POST /settings/test-gitlab
func (h *SettingsHandler) TestGitLab(w http.ResponseWriter, r *http.Request) {
	h.testConnection(w, r, "gitlab")
}

func (h *SettingsHandler) testConnection(w http.ResponseWriter, r *http.Request, providerType string) {
	var req struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.ReturnRequestError(w, "invalid request body")
		return
	}

	err := h.settingsService.TestConnectionDirect(r.Context(), providerType, req.URL, req.Token)
	if err != nil {
		rest.ReturnResponse(w, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"success": true,
	})
}
