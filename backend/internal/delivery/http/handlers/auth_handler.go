package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/middleware"
	identityApp "github.com/vulnex/vulnex/backend/internal/identity/application"
	identityDomain "github.com/vulnex/vulnex/backend/internal/identity/domain"
)

type AuthHandler struct {
	setupAdminUseCase *identityApp.SetupInitialAdminUseCase
	authUseCase       *identityApp.AuthenticateUserUseCase
	userRepo          identityDomain.UserRepository
}

func NewAuthHandler(
	setupAdmin *identityApp.SetupInitialAdminUseCase,
	authUseCase *identityApp.AuthenticateUserUseCase,
	userRepo identityDomain.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		setupAdminUseCase: setupAdmin,
		authUseCase:       authUseCase,
		userRepo:          userRepo,
	}
}

type SetupRequest struct {
	SetupToken  string `json:"setup_token"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (h *AuthHandler) SetupInitialAdmin(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := identityApp.SetupInitialAdminCommand{
		RawToken:    req.SetupToken,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	res, err := h.setupAdminUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "SETUP_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, res)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := identityApp.AuthenticateUserCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := h.authUseCase.Execute(r.Context(), cmd)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "invalid email or password")
		return
	}

	// For MVP token: user ID as Bearer token or signed session
	response := map[string]any{
		"token": res.UserID,
		"user": map[string]any{
			"id":           res.UserID,
			"email":        res.Email,
			"display_name": res.DisplayName,
		},
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	if userIDStr == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated session")
		return
	}

	userID, err := identityDomain.NewUserID(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user ID")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id":           user.ID().String(),
		"email":        user.Email().String(),
		"display_name": user.DisplayName(),
		"status":       string(user.Status()),
	})
}
