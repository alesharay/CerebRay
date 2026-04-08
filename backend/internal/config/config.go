package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	Env     string
	BaseURL string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Auth
	KeycloakIssuerURL    string
	KeycloakClientID     string
	KeycloakClientSecret string
	SessionSecret        string
	LocalUserEmail       string
	LocalUserName        string

	// AI
	AIEnabled            bool
	AnthropicAPIKey      string
	AIModel              string
	AIMonthlyTokenBudget int
}

func Load() (*Config, error) {
	// Load .env in development (ignore error if missing)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                 getStr("PORT", "8080"),
		Env:                  getStr("ENV", "development"),
		BaseURL:              getStr("BASE_URL", "http://localhost:5173"),
		DatabaseURL:          getStr("DATABASE_URL", ""),
		RedisURL:             getStr("REDIS_URL", "redis://localhost:6379"),
		KeycloakIssuerURL:    getStr("KEYCLOAK_ISSUER_URL", ""),
		KeycloakClientID:     getStr("KEYCLOAK_CLIENT_ID", "cerebray"),
		KeycloakClientSecret: getStr("KEYCLOAK_CLIENT_SECRET", ""),
		SessionSecret:        getStr("SESSION_SECRET", ""),
		LocalUserEmail:       getStr("LOCAL_USER_EMAIL", "admin@localhost"),
		LocalUserName:        getStr("LOCAL_USER_NAME", "Local Admin"),
		AIEnabled:            getBool("AI_ENABLED", true),
		AnthropicAPIKey:      getStr("ANTHROPIC_API_KEY", ""),
		AIModel:              getStr("AI_MODEL", "claude-sonnet-4-5-20250514"),
		AIMonthlyTokenBudget: getInt("AI_MONTHLY_TOKEN_BUDGET", 2000000),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsKeycloakEnabled() bool {
	return c.KeycloakIssuerURL != ""
}

func (c *Config) validate() error {
	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.SessionSecret == "" {
		missing = append(missing, "SESSION_SECRET")
	}
	if c.IsKeycloakEnabled() {
		if c.KeycloakClientID == "" {
			missing = append(missing, "KEYCLOAK_CLIENT_ID")
		}
		if c.KeycloakClientSecret == "" {
			missing = append(missing, "KEYCLOAK_CLIENT_SECRET")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func getStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
