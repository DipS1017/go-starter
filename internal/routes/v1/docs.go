package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/dba/internal/handlers"
)

func DocsRoutes(h *handlers.Handler, router *echo.Group) {
	router.GET("", h.ServeDocs)
	router.GET("/swagger.json", h.ServeSwagger)
}
