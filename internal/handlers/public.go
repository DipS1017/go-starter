package handlers

import (
	"github.com/labstack/echo/v4"
)

// Healthz returns a 200 OK with a basic JSON payload
func (h *Handler) Healthz(c echo.Context) error {
	status := map[string]string{"status": "ok"}
	return h.res.JSON(c, status)
}
