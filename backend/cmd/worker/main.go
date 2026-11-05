package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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

const version = "0.1.0-dev"

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
			processNextJob(ctx, workerID, scanJobQueue, scanRepo, nmapAdapter, nmapParser, zapAdapter, zapParser, ingestFindingsUC)
		}
	}
}

func processNextJob(
	ctx context.Context,
	workerID string,
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

		// Determine scanner from target or profile: web URLs use ZAP, hosts/IPs use Nmap
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			artifact, err := zapAdapter.ExecuteScan(ctx, scanDomain.ScanExecutionRequest{
				ScanID:      scan.ID(),
				TargetHost:  target,
				ProfileName: "zap-baseline",
				Timeout:     10 * time.Minute,
			})
			if err != nil {
				log.Printf("[VULNEX WORKER] ZAP execution failed on target %s: %v", target, err)
				continue
			}
			obs, err := zapParser.Parse(artifact.OutputReader, 50*1024*1024)
			if err != nil {
				log.Printf("[VULNEX WORKER] ZAP parse error on target %s: %v", target, err)
				continue
			}
			allObservations = append(allObservations, obs...)
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

	if len(dtos) > 0 {
		_, err := ingestUC.Execute(ctx, findingApp.IngestFindingsCommand{
			WorkspaceID:  scan.WorkspaceID().String(),
			ProjectID:    scan.ProjectID().String(),
			ScanID:       scan.ID().String(),
			AssetID:      scan.ProjectID().String() + "-asset",
			Observations: dtos,
			DiscoveredAt: time.Now().UTC(),
		})
		if err != nil {
			log.Printf("[VULNEX WORKER] Finding ingestion error: %v", err)
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
