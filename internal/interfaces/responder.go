package interfaces

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
)

type Responders interface {
	JSON(c echo.Context, payload any, opt ...dto.ResponderOptions) error
	JSONList(c echo.Context, payload any, totalPages int, code ...int) error
}
