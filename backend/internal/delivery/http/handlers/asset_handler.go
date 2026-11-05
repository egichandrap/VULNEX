package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	assetDomain "github.com/vulnex/vulnex/backend/internal/asset/domain"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
)

type AssetHandler struct {
	assetRepo assetDomain.AssetRepository
}

func NewAssetHandler(assetRepo assetDomain.AssetRepository) *AssetHandler {
	return &AssetHandler{assetRepo: assetRepo}
}

func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := assetDomain.NewProjectID(projectIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project ID format")
		return
	}

	assets, err := h.assetRepo.ListByProject(r.Context(), projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list assets")
		return
	}

	type AssetDTO struct {
		ID             string `json:"id"`
		ProjectID      string `json:"project_id"`
		AssetType      string `json:"asset_type"`
		CanonicalValue string `json:"canonical_value"`
		DisplayValue   string `json:"display_value"`
		Criticality    string `json:"criticality"`
		Status         string `json:"status"`
	}

	dtos := make([]AssetDTO, 0, len(assets))
	for _, a := range assets {
		dtos = append(dtos, AssetDTO{
			ID:             a.ID().String(),
			ProjectID:      a.ProjectID().String(),
			AssetType:      string(a.AssetType()),
			CanonicalValue: a.CanonicalValue().String(),
			DisplayValue:   a.DisplayValue(),
			Criticality:    string(a.Criticality()),
			Status:         string(a.Status()),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"assets": dtos})
}

type CreateAssetRequest struct {
	AssetType    string `json:"asset_type"`
	RawValue     string `json:"raw_value"`
	DisplayValue string `json:"display_value"`
	Criticality  string `json:"criticality"`
}

func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := assetDomain.NewProjectID(projectIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project ID format")
		return
	}

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	assetType := assetDomain.AssetType(req.AssetType)
	canonicalVal, err := assetDomain.NewCanonicalAssetValue(assetType, req.RawValue)
	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "INVALID_ASSET_VALUE", err.Error())
		return
	}

	crit := assetDomain.AssetCriticality(req.Criticality)
	if crit == "" {
		crit = assetDomain.CriticalityMedium
	}

	asset, err := domainAssetConstructor(projectID, assetType, canonicalVal, req.DisplayValue, crit)
	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
		return
	}

	if err := h.assetRepo.Save(r.Context(), asset); err != nil {
		httputil.WriteError(w, http.StatusConflict, "ASSET_EXISTS_OR_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"id":              asset.ID().String(),
		"canonical_value": asset.CanonicalValue().String(),
		"status":          string(asset.Status()),
	})
}

func domainAssetConstructor(
	projectID assetDomain.ProjectID,
	assetType assetDomain.AssetType,
	canonicalVal assetDomain.CanonicalAssetValue,
	displayValue string,
	crit assetDomain.AssetCriticality,
) (*assetDomain.Asset, error) {
	return assetDomain.NewAsset(projectID, assetType, canonicalVal, displayValue, crit)
}
