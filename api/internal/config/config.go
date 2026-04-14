package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv                            string
	Port                              string
	SupabaseJWTSecret                 string
	DatabaseURL                       string
	SupabaseURL                       string
	SupabaseServiceRoleKey            string
	SupabaseStorageBucket             string
	GeminiAPIKey                      string
	GeminiModel                       string
	IdempotencyCleanupIntervalSeconds int
	IdempotencyCleanupBatchSize       int64
	ChileCompraBaseURL                string
	ChileCompraAPIKey                 string
	ChileCompraTendersPath            string
	TenderScoreCacheTTLSeconds        int
}

func Load() Config {
	cfg := Config{
		AppEnv:                            getOrDefault("APP_ENV", "development"),
		Port:                              getOrDefault("PORT", "8080"),
		SupabaseJWTSecret:                 os.Getenv("SUPABASE_JWT_SECRET"),
		DatabaseURL:                       os.Getenv("DATABASE_URL"),
		SupabaseURL:                       os.Getenv("SUPABASE_URL"),
		SupabaseServiceRoleKey:            os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseStorageBucket:             getOrDefault("SUPABASE_STORAGE_BUCKET", "vault-items"),
		GeminiAPIKey:                      os.Getenv("GEMINI_API_KEY"),
		GeminiModel:                       getOrDefault("GEMINI_MODEL", "gemini-1.5-flash"),
		IdempotencyCleanupIntervalSeconds: getIntOrDefault("IDEMPOTENCY_CLEANUP_INTERVAL_SECONDS", 0),
		IdempotencyCleanupBatchSize:       int64(getIntOrDefault("IDEMPOTENCY_CLEANUP_BATCH_SIZE", 500)),
		ChileCompraBaseURL:                os.Getenv("CHILECOMPRA_BASE_URL"),
		ChileCompraAPIKey:                 os.Getenv("CHILECOMPRA_API_KEY"),
		ChileCompraTendersPath:            getOrDefault("CHILECOMPRA_TENDERS_PATH", "/servicios/v1/publico/licitaciones.json"),
		TenderScoreCacheTTLSeconds:        getIntOrDefault("TENDER_SCORE_CACHE_TTL_SECONDS", 900),
	}

	return cfg
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
