package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/db"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/handlers"
	// "github.com/webpoint-solutions-llc/go-starter/internal/middlewares"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/migrator"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/redisclient"
	v1 "github.com/webpoint-solutions-llc/go-starter/internal/routes/v1"
)

// ServerConfig holds configuration for server initialization
type ServerConfig struct {
	SkipMigrations bool
	SkipRedis      bool
}

// InitializeConfig initializes the application configuration
func InitializeConfig() error {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file. Checking runtime...")
	}

	config.Init()

	// Configure slog based on environment
	var logLevel slog.Level
	switch config.Cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add stack traces for errors
			if a.Key == slog.LevelKey {
				if err, ok := a.Value.Any().(error); ok {
					return slog.Attr{
						Key:   slog.LevelKey,
						Value: slog.StringValue(fmt.Sprintf("%+v", err)),
					}
				}
			}
			return a
		},
	}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if config.Cfg.AppEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	if warnings, err := config.Validate(); err != nil {
		return err
	} else {
		for _, msg := range warnings {
			slog.Warn(msg)
		}
	}

	return nil
}

// InitializeServices initializes external services like Redis and Elasticsearch
func InitializeServices(cfg *ServerConfig) error {
	if cfg == nil {
		cfg = &ServerConfig{}
	}

	db.OpenDbConnection(config.Cfg.PostgresqlURL)

	// Initialize Redis
	if !cfg.SkipRedis {
		redisclient.RedisConnect()
	}
	// Run migrations
	if !cfg.SkipMigrations {
		if err := migrator.StartMigrate(context.Background(), config.Cfg.PostgresqlURL, db.MigrationsFiles); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	return nil
}

// NewEchoServer creates and configures a new Echo server instance
func NewEchoServer() *echo.Echo {
	e := echo.New()

	e.Server.ReadTimeout = 300 * time.Second
	e.Server.WriteTimeout = 300 * time.Second
	e.Server.IdleTimeout = 120 * time.Second

	e.Use(middleware.BodyLimit("100M"))

	h := handlers.NewHandler()

	e.Validator = errorhandler.NewValidator()
	e.HTTPErrorHandler = errorhandler.CentralEchoErrorHandler

	e.Use(middleware.RequestID())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogError:  true,

		HandleError: false,

		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status >= 400 || v.Error != nil {
				slog.Warn("Request",
					"Method", v.Method,
					"URI", v.URI,
					"Status", v.Status,
					"Error", v.Error,
				)
				return nil
			}
			slog.Info("Request",
				"Method", v.Method,
				"URI", v.URI,
				"Status", v.Status,
			)
			return nil
		},
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.Recover())

	// CORS configuration
	// Add "x-api-key" to AllowHeaders if using API key authentication
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.Cfg.AllowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.Use(middleware.RequestID())

	// e.Use(middlewares.ApiKeyAuthMiddleware(
	// 	"/favicon.ico",
	// 	"/ws",
	// 	"/api/v1/docs",
	// 	"/api/v1/docs/swagger.json",
	// 	"/api/v1/public/healthz",
	// 	"/api/v1/public/stripe-webhook",
	// 	"/api/v1/auth/google/login",
	// 	"/api/v1/auth/apple/login",
	// 	"/api/v1/auth/email-verify",
	// ))

	// Load routes
	apiV1 := e.Group("/api/v1")
	v1.Load(h, apiV1)

	e.RouteNotFound("/*", func(c echo.Context) error {
		return errorhandler.ErrorNotFound("Route Not Found")
	})

	return e
}

// InitializeServer initializes the entire server with all dependencies
func InitializeServer(cfg *ServerConfig) (*echo.Echo, error) {
	// Initialize configuration
	if err := InitializeConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize services
	if err := InitializeServices(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Create Echo server
	e := NewEchoServer()

	return e, nil
}
