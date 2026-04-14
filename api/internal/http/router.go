package httpserver

import (
	"log"
	"net/http"
	"time"

	"nexus/api/internal/audit"
	"nexus/api/internal/chilecompra"
	"nexus/api/internal/companyprofile"
	"nexus/api/internal/http/handlers"
	"nexus/api/internal/http/middleware"
	"nexus/api/internal/idempotency"
	"nexus/api/internal/observability"
	"nexus/api/internal/storage"
	"nexus/api/internal/tenders"
	"nexus/api/internal/vault"
)

type RouterConfig struct {
	JWTSecret           string
	VaultStore          vault.Store
	Signer              storage.Signer
	Reader              storage.ObjectReader
	Extractor           vault.Extractor
	Audit               audit.Service
	Idempotency         idempotency.Service
	TendersStore        tenders.Store
	ChileCompraClient   chilecompra.Client
	CompanyProfileStore companyprofile.Store
	TenderScoreCache    tenders.ScoreCache
	TenderScoreCacheTTL time.Duration
	Metrics             *observability.Metrics
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", handlers.Liveness)
	mux.HandleFunc("GET /health/ready", handlers.Readiness)
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		if cfg.Metrics == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(cfg.Metrics.RenderPrometheus()))
	})

	authenticated := middleware.RequireAuth(middleware.AuthConfig{
		JWTSecret: cfg.JWTSecret,
	})
	if cfg.Signer == nil {
		cfg.Signer = storage.PlaceholderSigner{}
	}
	vaultHandler := handlers.VaultHandler{
		Store:       cfg.VaultStore,
		Signer:      cfg.Signer,
		Reader:      cfg.Reader,
		Extractor:   cfg.Extractor,
		Audit:       cfg.Audit,
		Idempotency: cfg.Idempotency,
		Metrics:     cfg.Metrics,
	}
	tendersHandler := handlers.TendersHandler{
		Store:         cfg.TendersStore,
		Client:        cfg.ChileCompraClient,
		Profile:       cfg.CompanyProfileStore,
		ScoreCache:    cfg.TenderScoreCache,
		ScoreCacheTTL: cfg.TenderScoreCacheTTL,
		Metrics:       cfg.Metrics,
	}
	companyProfileHandler := handlers.CompanyProfileHandler{
		Store:      cfg.CompanyProfileStore,
		ScoreCache: cfg.TenderScoreCache,
	}
	opsHandler := handlers.OpsHandler{
		Metrics: cfg.Metrics,
	}
	mux.Handle("GET /v1/protected", authenticated(http.HandlerFunc(handlers.ProtectedExample)))
	mux.Handle("POST /v1/vault/upload", authenticated(http.HandlerFunc(vaultHandler.Upload)))
	mux.Handle("POST /v1/vault/process", authenticated(http.HandlerFunc(vaultHandler.Process)))
	mux.Handle("POST /v1/vault/items/{id}/retry", authenticated(http.HandlerFunc(vaultHandler.RetryItem)))
	mux.Handle("GET /v1/vault/items", authenticated(http.HandlerFunc(vaultHandler.ListItems)))
	mux.Handle("GET /v1/vault/items/{id}", authenticated(http.HandlerFunc(vaultHandler.GetItem)))
	mux.Handle("GET /v1/vault/items/{id}/events", authenticated(http.HandlerFunc(vaultHandler.GetItemEvents)))
	mux.Handle("GET /v1/tenders/sync", authenticated(http.HandlerFunc(tendersHandler.Sync)))
	mux.Handle("GET /v1/tenders", authenticated(http.HandlerFunc(tendersHandler.List)))
	mux.Handle("GET /v1/tenders/{id}/score", authenticated(http.HandlerFunc(tendersHandler.Score)))
	mux.Handle("POST /v1/tenders/score/warmup", authenticated(http.HandlerFunc(tendersHandler.WarmupScoreCache)))
	mux.Handle("GET /v1/company/profile", authenticated(http.HandlerFunc(companyProfileHandler.Get)))
	mux.Handle("PUT /v1/company/profile", authenticated(http.HandlerFunc(companyProfileHandler.Upsert)))
	mux.Handle("GET /v1/ops/alerts", authenticated(http.HandlerFunc(opsHandler.Alerts)))

	return middleware.WithRequestID(withRequestLogging(mux, cfg.Metrics))
}

func withRequestLogging(next http.Handler, metrics *observability.Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		durationMs := time.Since(startedAt).Milliseconds()

		requestID := middleware.RequestIDFromContext(r.Context())
		log.Printf(
			"request_id=%s method=%s path=%s status=%d duration_ms=%d",
			requestID,
			r.Method,
			r.URL.Path,
			rec.statusCode,
			durationMs,
		)
		metrics.RecordHTTPRequest(r.Method, r.URL.Path, rec.statusCode, durationMs)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
