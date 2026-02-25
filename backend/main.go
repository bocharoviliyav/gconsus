package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gconsus/database"
	"gconsus/handler"
	"gconsus/lib/http/middleware"
	"gconsus/lib/logging"
	"gconsus/repository"
	"gconsus/service"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	_ "go.uber.org/automaxprocs"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string `env:"DB_HOST" envDefault:"localhost"`
	Port         int    `env:"DB_PORT" envDefault:"5432"`
	Name         string `env:"DB_NAME" envDefault:"git_analytics"`
	User         string `env:"DB_USER" envDefault:"git_analytics"`
	Password     string `env:"DB_PASSWORD" envDefault:"changeme123"`
	SSLMode      string `env:"DB_SSLMODE" envDefault:"disable"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	URL          string `env:"KEYCLOAK_URL" envDefault:"http://localhost:8090"`
	Realm        string `env:"KEYCLOAK_REALM" envDefault:"gconsus"`
	ClientID     string `env:"KEYCLOAK_CLIENT_ID" envDefault:"gconsus-backend"`
	ClientSecret string `env:"KEYCLOAK_CLIENT_SECRET"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port               int      `env:"SERVER_PORT" envDefault:"8000"`
	RequestSizeMax     int64    `env:"REQUEST_SIZE_MAX" envDefault:"1048576"` // 1MB
	ReadTimeout        int      `env:"READ_TIMEOUT" envDefault:"10"`
	WriteTimeout       int      `env:"WRITE_TIMEOUT" envDefault:"30"`
	IdleTimeout        int      `env:"IDLE_TIMEOUT" envDefault:"120"`
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost"`
}

// Config holds all application configuration
type Config struct {
	Database     DatabaseConfig
	Keycloak     KeycloakConfig
	Server       ServerConfig
	DebugLogging bool `env:"DEBUG_LOGGING" envDefault:"false"`
}

func main() {

	// Parse configuration
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logging.InitLogger(cfg.DebugLogging)
	slog.Info("Starting Git Analytics API", "port", cfg.Server.Port, "debug", cfg.DebugLogging)

	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Database connection established")

	// Initialize validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	syncRepo := repository.NewSyncRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, teamRepo, validate)
	teamService := service.NewTeamService(teamRepo, userRepo, validate)

	// Initialize aggregation service first
	aggregationService := service.NewAggregationService(activityRepo, metricsRepo, userRepo, teamRepo)

	// Initialize analytics service with all dependencies
	analyticsService := service.NewAnalyticsService(metricsRepo, activityRepo, userRepo, teamRepo, aggregationService)

	// Initialize settings and sync services
	settingsService := service.NewSettingsService(providerRepo, syncRepo)
	syncService := service.NewSyncService(activityRepo, syncRepo, userRepo, settingsService, aggregationService)

	// Initialize HTTP handlers
	handlerConfig := handler.Config{
		Port:           cfg.Server.Port,
		RequestSizeMax: cfg.Server.RequestSizeMax,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
	}

	httpServer, err := initHTTPServer(handlerConfig, cfg.Server.CORSAllowedOrigins, HTTPServices{
		UserService:        userService,
		TeamService:        teamService,
		AnalyticsService:   analyticsService,
		AggregationService: aggregationService,
		SettingsService:    settingsService,
		SyncService:        syncService,
	})
	if err != nil {
		slog.Error("Failed to initialize HTTP server", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	go func() {
		slog.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	slog.Info("Received interrupt signal, shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server forced to shutdown", "error", err)
	} else {
		slog.Info("HTTP server stopped gracefully")
	}
}

// initDatabase initializes the database connection
func initDatabase(cfg DatabaseConfig) (*database.DB, error) {
	dbConfig := database.Config{
		Host:         cfg.Host,
		Port:         cfg.Port,
		Database:     cfg.Name,
		User:         cfg.User,
		Password:     cfg.Password,
		SSLMode:      cfg.SSLMode,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// HTTPServices holds all services needed by HTTP handlers
type HTTPServices struct {
	UserService        *service.UserService
	TeamService        *service.TeamService
	AnalyticsService   *service.AnalyticsService
	AggregationService *service.AggregationService
	SettingsService    *service.SettingsService
	SyncService        *service.SyncService
}

// initHTTPServer initializes the HTTP server with all routes
func initHTTPServer(cfg handler.Config, corsOrigins []string, services HTTPServices) (*http.Server, error) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"git-analytics-api"}`))
	})

	// Root endpoint (exact match only -- no method prefix to avoid conflict with sub-routes)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Git Analytics API","version":"1.0.0","endpoints":["/health","/api/v1"]}`))
	})

	// API v1 routes
	apiV1 := http.NewServeMux()

	// User routes
	if services.UserService != nil {
		userHandler := handler.NewUsersHandler(services.UserService)
		apiV1.HandleFunc("POST /users", userHandler.CreateUser)
		apiV1.HandleFunc("GET /users/{id}", userHandler.GetUser)
		apiV1.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)
		apiV1.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser)
		apiV1.HandleFunc("GET /users", userHandler.ListUsers)
	}

	// Team routes
	if services.TeamService != nil {
		teamHandler := handler.NewTeamsHandler(services.TeamService)
		apiV1.HandleFunc("POST /teams", teamHandler.CreateTeam)
		apiV1.HandleFunc("GET /teams/{id}", teamHandler.GetTeam)
		apiV1.HandleFunc("PUT /teams/{id}", teamHandler.UpdateTeam)
		apiV1.HandleFunc("DELETE /teams/{id}", teamHandler.DeleteTeam)
		apiV1.HandleFunc("GET /teams", teamHandler.ListTeams)
		apiV1.HandleFunc("GET /teams/{id}/members", teamHandler.GetTeamMembers)
		apiV1.HandleFunc("POST /teams/{id}/members", teamHandler.AddTeamMember)
		apiV1.HandleFunc("DELETE /teams/{id}/members/{userId}", teamHandler.RemoveTeamMember)
		apiV1.HandleFunc("PATCH /teams/{id}/members/{userId}", teamHandler.UpdateMemberRole)
	}

	// Analytics routes
	if services.AnalyticsService != nil {
		analyticsHandler := handler.NewAnalyticsHandler(services.AnalyticsService)
		apiV1.HandleFunc("GET /analytics/users/{id}", analyticsHandler.GetUserAnalytics)
		apiV1.HandleFunc("GET /analytics/teams/{id}", analyticsHandler.GetTeamAnalytics)
		apiV1.HandleFunc("GET /analytics/leaderboard", analyticsHandler.GetLeaderboard)
		apiV1.HandleFunc("GET /analytics/teams/leaderboard", analyticsHandler.GetTeamLeaderboard)
		apiV1.HandleFunc("GET /analytics/dashboard", analyticsHandler.GetDashboard)
		apiV1.HandleFunc("GET /analytics/repositories", analyticsHandler.GetRepositoryAnalytics)
		apiV1.HandleFunc("GET /analytics/repositories/leaderboard", analyticsHandler.GetRepositoriesLeaderboard)
	}

	// Settings routes
	if services.SettingsService != nil {
		settingsHandler := handler.NewSettingsHandler(services.SettingsService)
		apiV1.HandleFunc("GET /settings/system", settingsHandler.GetSystemSettings)
		apiV1.HandleFunc("PUT /settings/system", settingsHandler.UpdateSystemSettings)
		apiV1.HandleFunc("GET /settings/providers", settingsHandler.ListProviders)
		apiV1.HandleFunc("POST /settings/providers", settingsHandler.CreateProvider)
		apiV1.HandleFunc("POST /settings/test-github", settingsHandler.TestGitHub)
		apiV1.HandleFunc("POST /settings/test-gitlab", settingsHandler.TestGitLab)
	}

	// Sync routes
	if services.SyncService != nil {
		syncHandler := handler.NewSyncHandler(services.SyncService)
		apiV1.HandleFunc("GET /sync/status", syncHandler.GetSyncStatus)
		apiV1.HandleFunc("POST /sync/trigger", syncHandler.TriggerSync)
		apiV1.HandleFunc("GET /sync/history", syncHandler.GetSyncHistory)
	}

	// Mount API v1 routes
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1))

	// Apply middleware
	wrapped := middleware.Use(
		mux,
		middleware.CORS(corsOrigins),
		middleware.RequestSizeMax(cfg.RequestSizeMax),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.Recoverer,
	)

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      wrapped,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	return server, nil
}
