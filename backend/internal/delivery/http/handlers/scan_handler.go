package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/middleware"
	scanApp "github.com/vulnex/vulnex/backend/internal/scanning/application"
	scanDomain "github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

type ScanHandler struct {
	requestScanUseCase *scanApp.RequestScanUseCase
	cancelScanUseCase  *scanApp.CancelScanUseCase
	scanRepo           scanDomain.ScanRepository
}

func NewScanHandler(
	requestScan *scanApp.RequestScanUseCase,
	cancelScan *scanApp.CancelScanUseCase,
	scanRepo scanDomain.ScanRepository,
) *ScanHandler {
	return &ScanHandler{
		requestScanUseCase: requestScan,
		cancelScanUseCase:  cancelScan,
		scanRepo:           scanRepo,
	}
}

type RequestScanPayload struct {
	WorkspaceID      string   `json:"workspace_id"`
	AuthorizationID  string   `json:"authorization_id"`
	ProfileVersionID string   `json:"profile_version_id"`
	IdempotencyKey   string   `json:"idempotency_key"`
	Targets          []string `json:"targets"`
}

func (h *ScanHandler) RequestScan(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectId")
	userIDStr := middleware.GetUserID(r)

	var payload RequestScanPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := scanApp.RequestScanCommand{
		WorkspaceID:      payload.WorkspaceID,
		ProjectID:        projectIDStr,
		AuthorizationID:  payload.AuthorizationID,
		ProfileVersionID: payload.ProfileVersionID,
		RequestedBy:      userIDStr,
		IdempotencyKey:   payload.IdempotencyKey,
		Targets:          payload.Targets,
	}

	res, err := h.requestScanUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "SCAN_REQUEST_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusAccepted, res)
}

func (h *ScanHandler) GetScan(w http.ResponseWriter, r *http.Request) {
	scanIDStr := chi.URLParam(r, "id")
	scanID, err := scanDomain.NewScanID(scanIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_SCAN_ID", "invalid scan ID")
		return
	}

	scan, err := h.scanRepo.FindByID(r.Context(), scanID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id":             scan.ID().String(),
		"workspace_id":   scan.WorkspaceID().String(),
		"project_id":     scan.ProjectID().String(),
		"status":         string(scan.Status()),
		"targets":        scan.TargetSnapshot(),
		"started_at":     scan.StartedAt(),
		"completed_at":   scan.CompletedAt(),
		"failure_reason": scan.FailureReason(),
		"created_at":     scan.CreatedAt(),
	})
}

type CancelScanPayload struct {
	Reason string `json:"reason"`
}

func (h *ScanHandler) CancelScan(w http.ResponseWriter, r *http.Request) {
	scanIDStr := chi.URLParam(r, "id")

	var payload CancelScanPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	cmd := scanApp.CancelScanCommand{
		ScanID: scanIDStr,
		Reason: payload.Reason,
	}

	err := h.cancelScanUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusConflict, "CANCEL_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"status": "CANCELLED"})
}
