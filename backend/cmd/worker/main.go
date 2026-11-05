package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	findingApp "github.com/vulnex/vulnex/backend/internal/findings/application"
	findingDomain "github.com/vulnex/vulnex/backend/internal/findings/domain"
	findingPostgres "github.com/vulnex/vulnex/backend/internal/findings/infrastructure/postgres"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
	scanDomain "github.com/vulnex/vulnex/backend/internal/scanning/domain"
	scanPostgres "github.com/vulnex/vulnex/backend/internal/scanning/infrastructure/postgres"
	"github.com/vulnex/vulnex/backend/internal/scanning/infrastructure/scanner"
)

const version = "1.0.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("VULNEX Scanner Worker daemon v%s\n", version)
		return
	}

	workerID := getEnv("WORKER_ID", "")
	if workerID == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "worker"
		}
		workerID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}

	dbDSN := getEnv("DATABASE_URL", "postgres://vulnex:vulnex_dev_password@localhost:5432/vulnex?sslmode=disable")
	pollInterval := 3 * time.Second

	log.Printf("[VULNEX WORKER] Starting daemon v%s (WorkerID: %s)...", version, workerID)

	db, err := postgres.OpenDB(postgres.Config{
		DSN:             dbDSN,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 15 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("[VULNEX WORKER] Database initialization error: %v", err)
	}
	defer db.Close()

	scanJobQueue := scanPostgres.NewScanJobQueue(db)
	scanRepo := scanPostgres.NewScanRepository(db)
	findingRepo := findingPostgres.NewFindingRepository(db)

	executor := scanner.NewOSCommandExecutor()
	nmapAdapter := scanner.NewNmapAdapter(executor)
	nmapParser := scanner.NewNmapParser()
	zapAdapter := scanner.NewZapAdapter(executor)
	zapParser := scanner.NewZapParser()

	scoringService := findingDomain.RiskScoringService{}
	ingestFindingsUC := findingApp.NewIngestFindingsUseCase(findingRepo, scoringService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[VULNEX WORKER] Shutdown signal received, terminating worker loop...")
		cancel()
	}()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("[VULNEX WORKER] Polling for scan tasks...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[VULNEX WORKER] Worker stopped cleanly.")
			return
		case <-ticker.C:
			processNextJob(ctx, workerID, db, scanJobQueue, scanRepo, nmapAdapter, nmapParser, zapAdapter, zapParser, ingestFindingsUC)
		}
	}
}

func ensureAssetID(ctx context.Context, db postgres.DBTX, projectID string, target string) (string, error) {
	canonical := target
	if u, err := url.Parse(target); err == nil && u.Host != "" {
		canonical = u.Host
	}

	var assetID string
	querySelect := `SELECT id FROM assets WHERE project_id = $1 AND canonical_value = $2 LIMIT 1`
	if err := db.QueryRowContext(ctx, querySelect, projectID, canonical).Scan(&assetID); err == nil {
		return assetID, nil
	}

	queryInsert := `
		INSERT INTO assets (id, project_id, asset_type, canonical_value, display_value, criticality, status)
		VALUES (gen_random_uuid(), $1, 'DOMAIN', $2, $2, 'MEDIUM', 'ACTIVE')
		ON CONFLICT (project_id, asset_type, canonical_value) DO UPDATE SET updated_at = clock_timestamp()
		RETURNING id;
	`
	if err := db.QueryRowContext(ctx, queryInsert, projectID, canonical).Scan(&assetID); err != nil {
		return "", fmt.Errorf("failed to ensure asset: %w", err)
	}
	return assetID, nil
}

func probeWebSecurity(ctx context.Context, targetURL string) []scanner.Observation {
	var obs []scanner.Observation
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; VULNEX-Scanner/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[VULNEX WORKER] Web probe failed on %s: %v", targetURL, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Security-Policy") == "" {
		obs = append(obs, scanner.Observation{
			Title:            "Content Security Policy (CSP) Header Not Set",
			Description:      "Content Security Policy is an effective defense against Cross-Site Scripting (XSS) and data injection attacks.",
			Severity:         "MEDIUM",
			Confidence:       "CONFIRMED",
			SourceScanner:    "web-baseline",
			VulnerabilityKey: "missing-csp",
			LocationKey:      targetURL,
		})
	}
	if strings.HasPrefix(targetURL, "https://") && resp.Header.Get("Strict-Transport-Security") == "" {
		obs = append(obs, scanner.Observation{
			Title:            "HTTP Strict Transport Security (HSTS) Missing",
			Description:      "The application does not enforce HTTPS via HSTS header, leaving users vulnerable to SSL stripping.",
			Severity:         "MEDIUM",
			Confidence:       "CONFIRMED",
			SourceScanner:    "web-baseline",
			VulnerabilityKey: "missing-hsts",
			LocationKey:      targetURL,
		})
	}
	if resp.Header.Get("X-Frame-Options") == "" {
		obs = append(obs, scanner.Observation{
			Title:            "X-Frame-Options Header Not Set (Clickjacking Risk)",
			Description:      "Missing X-Frame-Options header may allow attackers to frame the web application and perform clickjacking attacks.",
			Severity:         "MEDIUM",
			Confidence:       "CONFIRMED",
			SourceScanner:    "web-baseline",
			VulnerabilityKey: "missing-xfo",
			LocationKey:      targetURL,
		})
	}
	if resp.Header.Get("X-Content-Type-Options") == "" {
		obs = append(obs, scanner.Observation{
			Title:            "X-Content-Type-Options Header Missing (MIME Sniffing)",
			Description:      "Missing X-Content-Type-Options: nosniff header may allow browsers to perform MIME sniffing.",
			Severity:         "LOW",
			Confidence:       "CONFIRMED",
			SourceScanner:    "web-baseline",
			VulnerabilityKey: "missing-xcto",
			LocationKey:      targetURL,
		})
	}
	if s := resp.Header.Get("Server"); s != "" {
		obs = append(obs, scanner.Observation{
			Title:            fmt.Sprintf("Server Banner Information Disclosure (%s)", s),
			Description:      "Web server header exposes internal server software details to potential attackers.",
			Severity:         "LOW",
			Confidence:       "HIGH",
			SourceScanner:    "web-baseline",
			VulnerabilityKey: "info-disclosure-server",
			LocationKey:      targetURL,
		})
	}
	return obs
}

func processNextJob(
	ctx context.Context,
	workerID string,
	db postgres.DBTX,
	queue *scanPostgres.ScanJobQueue,
	scanRepo *scanPostgres.ScanRepository,
	nmapAdapter *scanner.NmapAdapter,
	nmapParser *scanner.NmapParser,
	zapAdapter *scanner.ZapAdapter,
	zapParser *scanner.ZapParser,
	ingestUC *findingApp.IngestFindingsUseCase,
) {
	job, err := queue.ClaimNext(ctx, workerID, 5*time.Minute)
	if err != nil {
		if errors.Is(err, scanPostgres.ErrNoJobAvailable) {
			// Normal idle polling state
			return
		}
		log.Printf("[VULNEX WORKER] Error claiming job: %v", err)
		return
	}

	log.Printf("[VULNEX WORKER] Claimed job %s for scan %s (attempt %d)", job.ID, job.ScanID, job.Attempt)

	scan, err := scanRepo.FindByID(ctx, job.ScanID)
	if err != nil {
		log.Printf("[VULNEX WORKER] Failed to load scan %s: %v", job.ScanID, err)
		_ = queue.Fail(ctx, job.ID, "scan record not found")
		return
	}

	now := time.Now().UTC()
	if err := scan.MarkRunning(now); err != nil {
		log.Printf("[VULNEX WORKER] Could not transition scan %s to RUNNING: %v", scan.ID(), err)
	}
	_ = scanRepo.Update(ctx, scan)

	// Execute scan targets
	var allObservations []scanner.Observation
	targets := scan.TargetSnapshot()

	for _, target := range targets {
		if ctx.Err() != nil {
			_ = scan.Cancel("worker shutdown interrupted scan")
			_ = scanRepo.Update(ctx, scan)
			_ = queue.Fail(ctx, job.ID, "worker shutdown")
			return
		}

		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			// Probe web security baseline
			webObs := probeWebSecurity(ctx, target)
			allObservations = append(allObservations, webObs...)

			// Also probe port scan on host if possible
			parsedHost := target
			if u, err := url.Parse(target); err == nil && u.Hostname() != "" {
				parsedHost = u.Hostname()
			}
			artifact, err := nmapAdapter.ExecuteScan(ctx, scanDomain.ScanExecutionRequest{
				ScanID:      scan.ID(),
				TargetHost:  parsedHost,
				ProfileName: "top-ports",
				Timeout:     3 * time.Minute,
			})
			if err == nil {
				obs, parseErr := nmapParser.Parse(artifact.OutputReader, 50*1024*1024)
				if parseErr == nil {
					allObservations = append(allObservations, obs...)
				}
			}
		} else {
			artifact, err := nmapAdapter.ExecuteScan(ctx, scanDomain.ScanExecutionRequest{
				ScanID:      scan.ID(),
				TargetHost:  target,
				ProfileName: "top-ports",
				Timeout:     5 * time.Minute,
			})
			if err != nil {
				log.Printf("[VULNEX WORKER] Nmap execution failed on target %s: %v", target, err)
				continue
			}
			obs, err := nmapParser.Parse(artifact.OutputReader, 50*1024*1024)
			if err != nil {
				log.Printf("[VULNEX WORKER] Nmap parse error on target %s: %v", target, err)
				continue
			}
			allObservations = append(allObservations, obs...)
		}
	}

	// Ingest observations as normalized findings
	if len(allObservations) > 0 && len(targets) > 0 {
		assetID, err := ensureAssetID(ctx, db, scan.ProjectID().String(), targets[0])
		if err != nil {
			log.Printf("[VULNEX WORKER] Failed to ensure asset: %v", err)
		} else {
			var dtos []findingApp.RawObservationDTO
			for _, o := range allObservations {
				dtos = append(dtos, findingApp.RawObservationDTO{
					Title:            o.Title,
					Description:      o.Description,
					Severity:         o.Severity,
					Confidence:       o.Confidence,
					SourceScanner:    o.SourceScanner,
					VulnerabilityKey: o.VulnerabilityKey,
					LocationKey:      o.LocationKey,
					AssetCriticality: "MEDIUM",
				})
			}
			_, err = ingestUC.Execute(ctx, findingApp.IngestFindingsCommand{
				WorkspaceID:  scan.WorkspaceID().String(),
				ProjectID:    scan.ProjectID().String(),
				ScanID:       scan.ID().String(),
				AssetID:      assetID,
				Observations: dtos,
				DiscoveredAt: time.Now().UTC(),
			})
			if err != nil {
				log.Printf("[VULNEX WORKER] Finding ingestion error: %v", err)
			}
		}
	}

	completedAt := time.Now().UTC()
	if err := scan.MarkCompleted(completedAt); err != nil {
		log.Printf("[VULNEX WORKER] Could not transition scan %s to COMPLETED: %v", scan.ID(), err)
	}
	_ = scanRepo.Update(ctx, scan)
	_ = queue.Complete(ctx, job.ID)

	log.Printf("[VULNEX WORKER] Successfully completed scan %s with %d observations", scan.ID(), len(allObservations))
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
