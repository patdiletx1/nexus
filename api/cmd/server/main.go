package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"nexus/api/internal/audit"
	"nexus/api/internal/chilecompra"
	"nexus/api/internal/companyprofile"
	"nexus/api/internal/config"
	httpserver "nexus/api/internal/http"
	"nexus/api/internal/idempotency"
	"nexus/api/internal/storage"
	"nexus/api/internal/tenders"
	"nexus/api/internal/vault"
)

func main() {
	cfg := config.Load()
	var signer storage.Signer = storage.PlaceholderSigner{}
	var reader storage.ObjectReader = storage.NoopObjectReader{}
	if cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != "" {
		signer = storage.SupabaseSigner{
			BaseURL:        cfg.SupabaseURL,
			ServiceRoleKey: cfg.SupabaseServiceRoleKey,
			Bucket:         cfg.SupabaseStorageBucket,
		}
		reader = storage.SupabaseObjectReader{
			BaseURL:        cfg.SupabaseURL,
			ServiceRoleKey: cfg.SupabaseServiceRoleKey,
			Bucket:         cfg.SupabaseStorageBucket,
		}
	}

	var extractor vault.Extractor = vault.NewSimulatedExtractor()
	if cfg.GeminiAPIKey != "" {
		extractor = vault.NewGeminiExtractor(cfg.GeminiAPIKey, cfg.GeminiModel)
	}

	store := vault.Store(vault.NewInMemoryStore())
	tendersStore := tenders.Store(tenders.NewInMemoryStore())
	companyProfileStore := companyprofile.Store(companyprofile.NewInMemoryStore())
	tenderScoreCache := tenders.ScoreCache(tenders.NewInMemoryScoreCache())
	var auditLogger audit.Service = audit.NoopLogger{}
	var idempotencyService idempotency.Service = idempotency.NoopService{}
	if cfg.DatabaseURL != "" {
		pool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Printf("postgres_connect_failed fallback=in_memory error=%v", err)
		} else {
			if pingErr := pool.Ping(context.Background()); pingErr != nil {
				log.Printf("postgres_ping_failed fallback=in_memory error=%v", pingErr)
				pool.Close()
			} else {
				store = vault.NewPostgresStore(pool)
				tendersStore = tenders.NewPostgresStore(pool)
				companyProfileStore = companyprofile.NewPostgresStore(pool)
				tenderScoreCache = tenders.NewPostgresScoreCache(pool)
				auditLogger = audit.NewPostgresLogger(pool)
				idempotencyService = idempotency.NewPostgresService(pool)
				defer pool.Close()
				log.Printf("vault_store=postgres")
			}
		}
	}

	var chileCompraClient chilecompra.Client = chilecompra.NoopClient{}
	if cfg.ChileCompraBaseURL != "" && cfg.ChileCompraAPIKey != "" {
		chileCompraClient = chilecompra.APIClient{
			BaseURL:     cfg.ChileCompraBaseURL,
			APIKey:      cfg.ChileCompraAPIKey,
			TendersPath: cfg.ChileCompraTendersPath,
		}
	}

	router := httpserver.NewRouter(httpserver.RouterConfig{
		JWTSecret:           cfg.SupabaseJWTSecret,
		VaultStore:          store,
		Signer:              signer,
		Reader:              reader,
		Extractor:           extractor,
		Audit:               auditLogger,
		Idempotency:         idempotencyService,
		TendersStore:        tendersStore,
		ChileCompraClient:   chileCompraClient,
		CompanyProfileStore: companyProfileStore,
		TenderScoreCache:    tenderScoreCache,
		TenderScoreCacheTTL: time.Duration(cfg.TenderScoreCacheTTLSeconds) * time.Second,
	})

	if cfg.IdempotencyCleanupIntervalSeconds > 0 {
		startIdempotencyCleanupJob(idempotencyService, time.Duration(cfg.IdempotencyCleanupIntervalSeconds)*time.Second, cfg.IdempotencyCleanupBatchSize)
	}

	log.Printf("starting_api env=%s port=%s", cfg.AppEnv, cfg.Port)
	if err := http.ListenAndServe(cfg.Address(), router); err != nil {
		log.Fatalf("server_stopped error=%v", err)
	}
}

func startIdempotencyCleanupJob(service idempotency.Service, interval time.Duration, batchSize int64) {
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			deleted, ok := service.CleanupExpired(context.Background(), batchSize)
			if !ok {
				continue
			}
			if deleted > 0 {
				log.Printf("idempotency_cleanup_deleted=%d", deleted)
			}
		}
	}()
}
