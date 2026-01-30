package interfaces

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
)

type Responders interface {
	JSON(c echo.Context, payload any, opt ...dto.ResponderOptions) error
	JSONList(c echo.Context, payload any, params utils.Params, count int64, code ...int) error
}
