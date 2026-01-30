package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/db"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/redisclient"
)

// Healthz returns a 200 OK with a basic JSON payload
func (h *Handler) Healthz(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()

	deps := map[string]string{}
	overall := "ok"
	code := http.StatusOK

	if db.Client == nil {
		deps["postgres"] = "not_initialized"
		overall = "degraded"
		code = http.StatusServiceUnavailable
	} else if err := db.Client.Ping(ctx); err != nil {
		deps["postgres"] = "down"
		overall = "degraded"
		code = http.StatusServiceUnavailable
	} else {
		deps["postgres"] = "ok"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" && redisclient.Client == nil {
		deps["redis"] = "disabled"
	} else if redisclient.Client == nil {
		deps["redis"] = "not_initialized"
		overall = "degraded"
		code = http.StatusServiceUnavailable
	} else if err := redisclient.Client.Ping(ctx).Err(); err != nil {
		deps["redis"] = "down"
		overall = "degraded"
		code = http.StatusServiceUnavailable
	} else {
		deps["redis"] = "ok"
	}

	status := map[string]any{
		"status":       overall,
		"dependencies": deps,
	}
	return c.JSON(code, status)
}
