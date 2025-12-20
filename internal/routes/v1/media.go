package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/handlers"
)

func MediaRoute(h *handlers.Handler, router *echo.Group) {
	router.POST("/upload", h.MediaUpload)
	router.DELETE("", h.MediaCleanup)
}
