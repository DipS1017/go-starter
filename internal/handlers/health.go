package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Healthz returns a 200 OK with a basic JSON payload
func Healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
