package handler

import (
	"net/http"
	"strconv"

	"gconsus/lib/http/rest"
	"gconsus/service"
)

// SyncHandler handles sync-related HTTP requests.
type SyncHandler struct {
	syncService *service.SyncService
}

// NewSyncHandler creates a new sync handler.
func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

// GetSyncStatus handles GET /sync/status
func (h *SyncHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.syncService.SyncStatus(r.Context())
	if err != nil {
		rest.ReturnServerError(w)
		return
	}
	rest.ReturnResponse(w, status)
}

// TriggerSync handles POST /sync/trigger
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	record, err := h.syncService.TriggerSync(r.Context())
	if err != nil {
		rest.ReturnRequestError(w, err.Error())
		return
	}
	rest.ReturnResponse(w, map[string]interface{}{
		"message": "sync started",
		"sync_id": record.ID,
	})
}

// GetSyncHistory handles GET /sync/history
func (h *SyncHandler) GetSyncHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	// Fetch a bit more than the page needs for total count
	records, err := h.syncService.SyncHistory(r.Context(), 100)
	if err != nil {
		rest.ReturnServerError(w)
		return
	}

	total := len(records)

	// Paginate
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}

	var items interface{}
	if offset < total {
		items = records[offset:end]
	} else {
		items = []interface{}{}
	}

	rest.ReturnResponse(w, map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}
