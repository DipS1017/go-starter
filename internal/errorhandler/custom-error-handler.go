package errorhandler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type failedResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

// CentralEchoErrorHandler now uses slog for structured logging
func CentralEchoErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return // Prevent duplicate JSON responses
	}

	// basic defaults
	code := http.StatusInternalServerError
	message := http.StatusText(code)
	var errPayload any

	// env/tracing
	requestID := c.Response().Header().Get("X-Request-Id")
	runtimeEnv := os.Getenv("RUNTIME_ENV")
	span := trace.SpanFromContext(c.Request().Context())

	// annotate span
	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("runtime.env", runtimeEnv),
	)
	span.SetStatus(codes.Error, err.Error())

	switch e := err.(type) {
	case HttpError:
		// log structured error
		slog.Error("HTTP error",
			slog.Int("status", e.Code),
			slog.String("method", c.Request().Method),
			slog.String("uri", c.Request().URL.RequestURI()),
			slog.String("caller", e.Caller),
			slog.String("request_id", requestID),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("error", err.Error()),
		)

		span.SetAttributes(attribute.String("error.caller", e.Caller))

		// TODO: Uncomment this when Slack integration is ready
		// for 5xx errors, notify Slack asynchronously
		// if e.Code >= 500 {
		// 	msg := slack.MessageBody{
		// 		Message:     err.Error(),
		// 		RequestId:   requestID,
		// 		ServiceName: serviceName + " " + runtimeEnv,
		// 		Code:        fmt.Sprintf("%d", e.Code),
		// 		Caller:      e.Caller,
		// 	}
		// 	go func() {
		// 		if sendErr := slack.SendSlackMessage(msg); sendErr != nil {
		// 			slog.Error("slack notification failed",
		// 				slog.Any("error", sendErr),
		// 			)
		// 		}
		// 	}()
		// }

		if e.Code < 500 {
			code = e.Code

			// Drill into e.Message if it’s a map, pull out the first error
			switch m := e.Message.(type) {
			case map[string]string:
				for _, msgStr := range m {
					message = msgStr
					break
				}
			case map[string][]string:
				for _, arr := range m {
					if len(arr) > 0 {
						message = arr[0]
						break
					}
				}
			case string:
				message = m
			case error:
				message = m.Error()
			default:
				// fallback for any other type
				message = fmt.Sprintf("%v", m)
			}

			// clear the payload so "error" is omitted in JSON
			errPayload = nil
		}

	default:
		// unexpected errors
		slog.Error("unexpected error",
			slog.String("method", c.Request().Method),
			slog.String("uri", c.Request().URL.RequestURI()),
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
		)
	}

	// send response
	if c.Request().Method == http.MethodHead {
		err = c.NoContent(code)
	} else {
		msg := failedResponse{Success: false, Message: message, Error: errPayload}
		err = c.JSON(code, msg)
	}
	if err != nil {
		slog.Error("failed to send error response", slog.Any("error", err))
	}
}

// CustomRequestLoggerConfig now emits Info‐level logs via slog
func CustomRequestLoggerConfig() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: false,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// only log successful requests here
			if v.Error == nil {
				traceID := trace.SpanFromContext(c.Request().Context()).SpanContext().TraceID().String()
				slog.Info("request completed",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("trace_id", traceID),
				)
			}
			return nil
		},
	})
}
