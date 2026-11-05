package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/middleware"
	identityApp "github.com/vulnex/vulnex/backend/internal/identity/application"
	identityDomain "github.com/vulnex/vulnex/backend/internal/identity/domain"
)

type AuthHandler struct {
	setupAdminUseCase     *identityApp.SetupInitialAdminUseCase
	authUseCase           *identityApp.AuthenticateUserUseCase
	changePasswordUseCase *identityApp.ChangePasswordUseCase
	userRepo              identityDomain.UserRepository
}

func NewAuthHandler(
	setupAdmin *identityApp.SetupInitialAdminUseCase,
	authUseCase *identityApp.AuthenticateUserUseCase,
	changePasswordUseCase *identityApp.ChangePasswordUseCase,
	userRepo identityDomain.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		setupAdminUseCase:     setupAdmin,
		authUseCase:           authUseCase,
		changePasswordUseCase: changePasswordUseCase,
		userRepo:              userRepo,
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

	emailStr := strings.TrimSpace(req.Email)
	if emailStr != "" && !strings.Contains(emailStr, "@") {
		emailStr = emailStr + "@vulnex.local"
	}

	cmd := identityApp.AuthenticateUserCommand{
		Email:    emailStr,
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

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r)
	if userIDStr == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated session")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON payload")
		return
	}

	cmd := identityApp.ChangePasswordCommand{
		UserID:          userIDStr,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.changePasswordUseCase.Execute(r.Context(), cmd); err != nil {
		switch err {
		case identityApp.ErrIncorrectCurrentPassword:
			httputil.WriteError(w, http.StatusUnauthorized, "INCORRECT_PASSWORD", err.Error())
		case identityApp.ErrPasswordEmpty:
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_PASSWORD", err.Error())
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "PASSWORD_UPDATE_FAILED", err.Error())
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "password updated successfully",
	})
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
