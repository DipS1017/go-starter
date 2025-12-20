package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/handlers"
)

func PublicRoutes(h *handlers.Handler, router *echo.Group) {
	router.GET("/healthz", h.Healthz)

	router.POST("/stripe-webhook", h.StripeWebhook)
	router.DELETE("/media-cleanup", h.MediaCleanup)
}
