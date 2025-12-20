package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/pkg/docs"
)

func (h *Handler) ServeDocs(c echo.Context) error {
	data, err := docs.IndexDoc.ReadFile("index.html")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.HTMLBlob(http.StatusOK, data)
}

func (h *Handler) ServeSwagger(c echo.Context) error {
	data, err := docs.SwaggerDoc.ReadFile("swagger.json")
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSONBlob(http.StatusOK, data)
}
