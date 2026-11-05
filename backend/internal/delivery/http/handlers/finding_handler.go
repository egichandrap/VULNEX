package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	findingDomain "github.com/vulnex/vulnex/backend/internal/findings/domain"
)

type FindingHandler struct {
	findingRepo findingDomain.FindingRepository
}

func NewFindingHandler(findingRepo findingDomain.FindingRepository) *FindingHandler {
	return &FindingHandler{findingRepo: findingRepo}
}

func (h *FindingHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := findingDomain.NewProjectID(projectIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "invalid project ID format")
		return
	}

	findings, err := h.findingRepo.ListByProject(r.Context(), projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list findings")
		return
	}

	type FindingDTO struct {
		ID            string    `json:"id"`
		WorkspaceID   string    `json:"workspace_id"`
		ProjectID     string    `json:"project_id"`
		ScanID        string    `json:"scan_id"`
		AssetID       string    `json:"asset_id"`
		Fingerprint   string    `json:"fingerprint"`
		Title         string    `json:"title"`
		Description   string    `json:"description"`
		Severity      string    `json:"severity"`
		Confidence    string    `json:"confidence"`
		Status        string    `json:"status"`
		SourceScanner string    `json:"source_scanner"`
		FirstSeenAt   time.Time `json:"first_seen_at"`
		LastSeenAt    time.Time `json:"last_seen_at"`
		RiskScore     float64   `json:"risk_score"`
	}

	dtos := make([]FindingDTO, 0, len(findings))
	for _, f := range findings {
		dtos = append(dtos, FindingDTO{
			ID:            f.ID().String(),
			WorkspaceID:   f.WorkspaceID().String(),
			ProjectID:     f.ProjectID().String(),
			ScanID:        f.ScanID().String(),
			AssetID:       f.AssetID().String(),
			Fingerprint:   f.Fingerprint().String(),
			Title:         f.Title(),
			Description:   f.Description(),
			Severity:      string(f.Severity()),
			Confidence:    string(f.Confidence()),
			Status:        string(f.Status()),
			SourceScanner: f.SourceScanner(),
			FirstSeenAt:   f.FirstSeenAt(),
			LastSeenAt:    f.LastSeenAt(),
			RiskScore:     f.RiskScore(),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"findings": dtos})
}

type TriageFindingPayload struct {
	Status        string `json:"status"` // "RESOLVED", "ACCEPTED_RISK", "FALSE_POSITIVE"
	Justification string `json:"justification"`
}

func (h *FindingHandler) TriageFinding(w http.ResponseWriter, r *http.Request) {
	findingIDStr := chi.URLParam(r, "id")
	findingID, err := findingDomain.NewFindingID(findingIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_FINDING_ID", "invalid finding ID")
		return
	}

	var payload TriageFindingPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	finding, err := h.findingRepo.FindByID(r.Context(), findingID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "FINDING_NOT_FOUND", "finding not found")
		return
	}

	now := time.Now().UTC()
	switch payload.Status {
	case "RESOLVED":
		err = finding.Resolve(now)
	case "ACCEPTED_RISK":
		err = finding.AcceptRisk(payload.Justification, now)
	case "FALSE_POSITIVE":
		err = finding.MarkFalsePositive(payload.Justification, now)
	default:
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_STATUS", "invalid finding status for triage")
		return
	}

	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "TRIAGE_FAILED", err.Error())
		return
	}

	if err := h.findingRepo.Update(r.Context(), finding); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update finding state")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id":     finding.ID().String(),
		"status": string(finding.Status()),
	})
}
