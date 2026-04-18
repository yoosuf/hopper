package regions

import (
	"net/http"

	"github.com/crewdigital/hopper/internal/platform/httpx"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for region operations
type Handler struct {
	regionService *Service
}

// NewHandler creates a new regions handler
func NewHandler(regionService *Service) *Handler {
	return &Handler{
		regionService: regionService,
	}
}

// ListRegions handles listing all active regions
func (h *Handler) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.regionService.ListRegions(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list regions", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, regions)
}

// GetRegion handles getting a specific region
func (h *Handler) GetRegion(w http.ResponseWriter, r *http.Request) {
	regionIDStr := r.URL.Query().Get("id")
	regionID, err := uuid.Parse(regionIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REGION_ID", "Invalid region ID", nil)
		return
	}

	region, err := h.regionService.GetRegion(r.Context(), regionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get region", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, region)
}

// GetRegionConfig handles getting region configuration
func (h *Handler) GetRegionConfig(w http.ResponseWriter, r *http.Request) {
	regionIDStr := r.URL.Query().Get("id")
	regionID, err := uuid.Parse(regionIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REGION_ID", "Invalid region ID", nil)
		return
	}

	config, err := h.regionService.GetRegionConfig(r.Context(), regionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get region config", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, config)
}
