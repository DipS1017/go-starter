package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/dba/internal/handlers"
	"github.com/webpoint-solutions-llc/dba/internal/middlewares"
)

func Load(h *handlers.Handler, router *echo.Group) {
	PublicRoutes(h, router.Group("/public"))
	AuthRoutes(h, router.Group("/auth"))
	DocsRoutes(h, router.Group("/docs"))

	UserRoutes(h, router.Group("/user", middlewares.AuthMiddleware("USER", "ADMIN")))
	MediaRoute(h, router.Group("/media", middlewares.AuthMiddleware("USER", "ADMIN")))
}
