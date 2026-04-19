package regions

import (
	"net/http"

	"github.com/yoosuf/hopper/internal/platform/httpx"
	"github.com/yoosuf/hopper/internal/platform/middleware"
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
// @Summary List all regions
// @Description Retrieve all active regions
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Region "Regions retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /regions [get]
func (h *Handler) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.regionService.ListRegions(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list regions", nil)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, regions)
}

// GetRegion handles getting a specific region
// @Summary Get region by ID
// @Description Retrieve a specific region by its ID
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Region ID"
// @Success 200 {object} Region "Region retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid region ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Region not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /regions/{id} [get]
func (h *Handler) GetRegion(w http.ResponseWriter, r *http.Request) {
	regionIDStr := middleware.URLParam(r, "id")
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
// @Summary Get region configuration
// @Description Retrieve the configuration for a specific region
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Region ID"
// @Success 200 {object} RegionConfig "Region configuration retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid region ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Region not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /regions/config [get]
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
