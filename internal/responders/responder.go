package responders

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/interfaces"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
)

type res struct {
	logger *slog.Logger
}

func NewResponder(logger *slog.Logger) interfaces.Responders {
	return &res{logger}
}

type successResponse struct {
	Success bool   `json:"success"`
	Payload any    `json:"payload"`
	Message string `json:"message"`
}

type successListResponse struct {
	Success     bool  `json:"success"`
	Payload     any   `json:"payload"`
	CurrentPage int32 `json:"current_page,omitempty"`
	TotalPages  int32 `json:"total_pages"`
	TotalCount  int64 `json:"total_count,omitempty"`
	HasNext     bool  `json:"has_next,omitempty"`
}

func (r *res) JSON(c echo.Context, payload any, opt ...dto.ResponderOptions) error {
	message := "Success"
	status := 200
	if len(opt) > 0 {

		if opt[0].Code != 0 {
			status = opt[0].Code
		}
		if opt[0].Message != "" {
			message = opt[0].Message
		}

	}
	return c.JSON(status, successResponse{Success: true, Payload: payload, Message: message})
}

func (r *res) JSONList(c echo.Context, payload any, params utils.Params, count int64, code ...int) error {
	status := 200
	if len(code) > 0 {
		status = code[0]
	}
	return c.JSON(status, successListResponse{
		Success:     true,
		Payload:     payload,
		TotalPages:  utils.GetTotalPages(count, params.Limit),
		TotalCount:  count,
		CurrentPage: params.CurrentPage,
		HasNext:     utils.GetHasNext(params.Offset, params.Limit, int32(count)),
	})
}
