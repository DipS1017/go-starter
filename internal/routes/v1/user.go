package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/handlers"
)

func UserRoutes(h *handlers.Handler, router *echo.Group) {
	router.POST("/upload-profile", h.UploadProfileImage)

	router.GET("/me", h.Me)
}
