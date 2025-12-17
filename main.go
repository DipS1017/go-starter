package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/webpoint-solutions-llc/dba/cmd"
	"github.com/webpoint-solutions-llc/dba/internal/config"
)

// @title          dba API
// @version         1.0
// @description     RESTful API for dba.

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token in the format: Bearer <token>

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-api-key
// @description Enter your API key in the format: x-api-key
func main() {
	// Initialize server with configuration
	e, err := cmd.InitializeServer(&cmd.ServerConfig{
		SkipMigrations: false,
		SkipRedis:      false,
	})
	if err != nil {
		slog.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}
	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err := e.Start(":" + config.Cfg.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
