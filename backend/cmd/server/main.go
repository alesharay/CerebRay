package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/ai"
	"github.com/aray/cerebray/backend/internal/auth"
	"github.com/aray/cerebray/backend/internal/config"
	"github.com/aray/cerebray/backend/internal/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// Logger setup
	if cfg.IsProduction() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Str("env", cfg.Env).Msg("starting cerebray")

	ctx := context.Background()

	// PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to postgres")
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatal().Err(err).Msg("pinging postgres")
	}
	log.Info().Msg("postgres connected")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing redis URL")
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("pinging redis")
	}
	defer redisClient.Close()
	log.Info().Msg("redis connected")

	// sqlc queries
	queries := sqlc.New(pool)

	// Session store
	sessions := auth.NewSessionStore(redisClient)

	// User upsert function (bridges auth to DB without circular deps)
	userUpsertFn := func(ctx context.Context, oidcSubject, email, name, avatar string) (int64, error) {
		// Try to find existing user
		existing, err := queries.GetUserByOIDCSubject(ctx, oidcSubject)
		if err == nil {
			// Update existing user
			updated, err := queries.UpdateUser(ctx, sqlc.UpdateUserParams{
				ID:        existing.ID,
				Email:     email,
				Name:      name,
				AvatarUrl: avatar,
			})
			if err != nil {
				return 0, err
			}
			return updated.ID, nil
		}
		// Create new user
		user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			OidcSubject: oidcSubject,
			Email:       email,
			Name:        name,
			AvatarUrl:   avatar,
		})
		if err != nil {
			return 0, err
		}
		return user.ID, nil
	}

	// Auth handler
	var oauthCfg *oauth2.Config
	if cfg.IsKeycloakEnabled() {
		oauthCfg = auth.NewKeycloakOAuthConfig(
			cfg.KeycloakIssuerURL,
			cfg.KeycloakClientID,
			cfg.KeycloakClientSecret,
			cfg.BaseURL+"/auth/callback",
		)
	}

	authHandler := auth.NewHandler(auth.HandlerConfig{
		OAuthConfig:  oauthCfg,
		IssuerURL:    cfg.KeycloakIssuerURL,
		Sessions:     sessions,
		UpsertUser:   userUpsertFn,
		BaseURL:      cfg.BaseURL,
		SecureCookie: cfg.IsProduction(),
		LocalMode:    !cfg.IsKeycloakEnabled(),
		LocalEmail:   cfg.LocalUserEmail,
		LocalName:    cfg.LocalUserName,
	})

	// AI provider (conditional on config)
	var aiProvider ai.Provider
	if cfg.AIEnabled && cfg.AnthropicAPIKey != "" {
		aiProvider = ai.NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.AIModel)
		log.Info().Str("model", cfg.AIModel).Msg("anthropic AI provider enabled")
	} else {
		log.Warn().Msg("AI provider disabled (no API key or AI_ENABLED=false)")
	}

	// Handlers
	noteHandlers := handlers.NewNoteHandlers(queries)
	tagHandlers := handlers.NewTagHandlers(queries)
	connHandlers := handlers.NewConnectionHandlers(queries)
	dashHandlers := handlers.NewDashboardHandlers(queries)
	glossaryHandlers := handlers.NewGlossaryHandlers(queries)
	convoHandlers := handlers.NewConversationHandlers(queries)
	chatHandlers := handlers.NewChatHandlers(queries, aiProvider)

	// Router
	router := buildRouter(RouterDeps{
		AuthHandler:   authHandler,
		Sessions:      sessions,
		Notes:         noteHandlers,
		Tags:          tagHandlers,
		Connections:   connHandlers,
		Dashboard:     dashHandlers,
		Glossary:      glossaryHandlers,
		Conversations: convoHandlers,
		Chat:          chatHandlers,
		AllowOrigin:   cfg.BaseURL,
	})

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // longer for SSE streaming
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", cfg.Port).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-done
	log.Info().Msg("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("server stopped")
}
