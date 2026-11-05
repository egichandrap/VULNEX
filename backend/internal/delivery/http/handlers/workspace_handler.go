package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/middleware"
	workspaceApp "github.com/vulnex/vulnex/backend/internal/workspace/application"
	workspaceDomain "github.com/vulnex/vulnex/backend/internal/workspace/domain"
)

type WorkspaceHandler struct {
	createWorkspaceUseCase *workspaceApp.CreateWorkspaceUseCase
	updatePolicyUseCase    *workspaceApp.UpdateWorkspacePolicyUseCase
	workspaceRepo          workspaceDomain.WorkspaceRepository
}

func NewWorkspaceHandler(
	createWs *workspaceApp.CreateWorkspaceUseCase,
	updatePolicy *workspaceApp.UpdateWorkspacePolicyUseCase,
	repo workspaceDomain.WorkspaceRepository,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		createWorkspaceUseCase: createWs,
		updatePolicyUseCase:    updatePolicy,
		workspaceRepo:          repo,
	}
}

func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	userID, err := workspaceDomain.NewUserID(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user ID")
		return
	}

	workspaces, err := h.workspaceRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list workspaces")
		return
	}

	type WorkspaceDTO struct {
		ID                      string `json:"id"`
		Name                    string `json:"name"`
		OwnerUserID             string `json:"owner_user_id"`
		AllowPrivateNetworkScan bool   `json:"allow_private_network_scan"`
		Status                  string `json:"status"`
	}

	dtos := make([]WorkspaceDTO, 0, len(workspaces))
	for _, ws := range workspaces {
		dtos = append(dtos, WorkspaceDTO{
			ID:                      ws.ID().String(),
			Name:                    ws.Name(),
			OwnerUserID:             ws.OwnerUserID().String(),
			AllowPrivateNetworkScan: ws.AllowPrivateNetworkScan(),
			Status:                  string(ws.Status()),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"workspaces": dtos})
}

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)

	var req CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := workspaceApp.CreateWorkspaceCommand{
		Name:        req.Name,
		OwnerUserID: userIDStr,
	}

	res, err := h.createWorkspaceUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, res)
}

type UpdatePolicyRequest struct {
	AllowPrivateNetworkScan bool   `json:"allow_private_network_scan"`
	Justification           string `json:"justification"`
}

func (h *WorkspaceHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "id")
	userIDStr := middleware.GetUserID(r)

	var req UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := workspaceApp.UpdateWorkspacePolicyCommand{
		WorkspaceID:             workspaceIDStr,
		ActorUserID:             userIDStr,
		AllowPrivateNetworkScan: req.AllowPrivateNetworkScan,
		Justification:           req.Justification,
	}

	err := h.updatePolicyUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusForbidden, "POLICY_UPDATE_FORBIDDEN", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"status": "POLICY_UPDATED"})
}
