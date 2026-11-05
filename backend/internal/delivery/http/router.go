package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/handlers"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/httputil"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/middleware"
)

type RouterConfig struct {
	TrustedProxies []string
	AuthHandler    *handlers.AuthHandler
	Workspace      *handlers.WorkspaceHandler
	Asset          *handlers.AssetHandler
	Scan           *handlers.ScanHandler
	Finding        *handlers.FindingHandler
	Authenticator  middleware.Authenticator
}

// NewRouter constructs and configures the Chi router.
func NewRouter(cfg RouterConfig) (http.Handler, error) {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// CORS policy
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Enforce trusted proxy CIDR evaluation (ADR-014)
	trustedProxyMw, err := middleware.NewTrustedProxyMiddleware(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}
	r.Use(trustedProxyMw.Handler)

	// Liveness & readiness endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	authMw := middleware.NewAuthMiddleware(cfg.Authenticator)

	r.Route("/api/v1", func(v1 chi.Router) {
		// Public Authentication routes
		v1.Post("/auth/setup", cfg.AuthHandler.SetupInitialAdmin)
		v1.Post("/auth/login", cfg.AuthHandler.Login)

		// Authenticated routes
		v1.Group(func(auth chi.Router) {
			auth.Use(authMw.RequireAuth)

			auth.Get("/auth/me", cfg.AuthHandler.GetMe)

			// Workspaces
			auth.Get("/workspaces", cfg.Workspace.ListWorkspaces)
			auth.Post("/workspaces", cfg.Workspace.CreateWorkspace)
			auth.Patch("/workspaces/{id}/network-policy", cfg.Workspace.UpdatePolicy)

			// Assets
			auth.Get("/projects/{projectId}/assets", cfg.Asset.ListAssets)
			auth.Post("/projects/{projectId}/assets", cfg.Asset.CreateAsset)

			// Scans
			auth.Post("/projects/{projectId}/scans", cfg.Scan.RequestScan)
			auth.Get("/scans/{id}", cfg.Scan.GetScan)
			auth.Post("/scans/{id}/cancel", cfg.Scan.CancelScan)

			// Findings
			auth.Get("/projects/{projectId}/findings", cfg.Finding.ListFindings)
			auth.Patch("/findings/{id}", cfg.Finding.TriageFinding)
		})
	})

	return r, nil
}
