package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	assetPostgres "github.com/vulnex/vulnex/backend/internal/asset/infrastructure/postgres"
	authPostgres "github.com/vulnex/vulnex/backend/internal/authorization/infrastructure/postgres"
	httpDelivery "github.com/vulnex/vulnex/backend/internal/delivery/http"
	"github.com/vulnex/vulnex/backend/internal/delivery/http/handlers"
	findingPostgres "github.com/vulnex/vulnex/backend/internal/findings/infrastructure/postgres"
	identityApp "github.com/vulnex/vulnex/backend/internal/identity/application"
	"github.com/vulnex/vulnex/backend/internal/identity/infrastructure/crypto"
	identityPostgres "github.com/vulnex/vulnex/backend/internal/identity/infrastructure/postgres"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
	scanApp "github.com/vulnex/vulnex/backend/internal/scanning/application"
	"github.com/vulnex/vulnex/backend/internal/scanning/infrastructure/authgate"
	scanPostgres "github.com/vulnex/vulnex/backend/internal/scanning/infrastructure/postgres"
	workspaceApp "github.com/vulnex/vulnex/backend/internal/workspace/application"
	workspacePostgres "github.com/vulnex/vulnex/backend/internal/workspace/infrastructure/postgres"
)

const version = "1.0.0"

type simpleAuthenticator struct {
	userRepo *identityPostgres.UserRepository
}

func (a *simpleAuthenticator) ValidateToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("VULNEX API Server v%s\n", version)
		return
	}

	port := getEnv("PORT", "8080")
	dbDSN := getEnv("DATABASE_URL", "postgres://vulnex:vulnex_dev_password@localhost:5432/vulnex?sslmode=disable")
	trustedProxiesEnv := getEnv("TRUSTED_PROXIES", "127.0.0.1/32,::1/128,172.16.0.0/12,10.0.0.0/8")
	trustedProxies := strings.Split(trustedProxiesEnv, ",")

	log.Printf("[VULNEX API] Initializing v%s on port %s...", version, port)

	// Open DB connection pool
	db, err := postgres.OpenDB(postgres.Config{
		DSN:             dbDSN,
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 15 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("[VULNEX API] Database initialization error: %v", err)
	}
	defer db.Close()

	// Instantiate infrastructure repositories
	userRepo := identityPostgres.NewUserRepository(db)
	tokenRepo := identityPostgres.NewSetupTokenRepository(db)
	workspaceRepo := workspacePostgres.NewWorkspaceRepository(db)
	assetRepo := assetPostgres.NewAssetRepository(db)
	authRepo := authPostgres.NewAuthorizationRepository(db)
	scanRepo := scanPostgres.NewScanRepository(db)
	scanJobQueue := scanPostgres.NewScanJobQueue(db)
	findingRepo := findingPostgres.NewFindingRepository(db)
	hasher := crypto.NewBcryptPasswordHasher(12)

	// Instantiate use cases
	setupAdminUC := identityApp.NewSetupInitialAdminUseCase(userRepo, tokenRepo, hasher)
	authUC := identityApp.NewAuthenticateUserUseCase(userRepo, hasher)
	createWsUC := workspaceApp.NewCreateWorkspaceUseCase(workspaceRepo)
	updatePolicyUC := workspaceApp.NewUpdateWorkspacePolicyUseCase(workspaceRepo)
	scanAuthGate := authgate.NewAuthorizationAdapter(authRepo)
	requestScanUC := scanApp.NewRequestScanUseCase(scanRepo, scanJobQueue, scanAuthGate)
	cancelScanUC := scanApp.NewCancelScanUseCase(scanRepo)

	// Instantiate HTTP handlers
	authHandler := handlers.NewAuthHandler(setupAdminUC, authUC, userRepo)
	workspaceHandler := handlers.NewWorkspaceHandler(createWsUC, updatePolicyUC, workspaceRepo)
	assetHandler := handlers.NewAssetHandler(assetRepo)
	scanHandler := handlers.NewScanHandler(requestScanUC, cancelScanUC, scanRepo)
	findingHandler := handlers.NewFindingHandler(findingRepo)

	// Build router
	router, err := httpDelivery.NewRouter(httpDelivery.RouterConfig{
		TrustedProxies: trustedProxies,
		AuthHandler:    authHandler,
		Workspace:      workspaceHandler,
		Asset:          assetHandler,
		Scan:           scanHandler,
		Finding:        findingHandler,
		Authenticator:  &simpleAuthenticator{userRepo: userRepo},
	})
	if err != nil {
		log.Fatalf("[VULNEX API] Router configuration error: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[VULNEX API] Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[VULNEX API] Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("[VULNEX API] Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[VULNEX API] Forced shutdown: %v", err)
	}

	log.Println("[VULNEX API] Server stopped.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
