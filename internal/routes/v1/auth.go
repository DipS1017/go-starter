package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/handlers"
	"github.com/webpoint-solutions-llc/go-starter/internal/middlewares"
)

func AuthRoutes(h *handlers.Handler, router *echo.Group) {
	router.GET("/google/login", h.GoogleLogin)
	router.POST("/google/callback", h.GoogleCodeExchange)
	router.GET("/apple/login", h.AppleLogin)
	router.POST("/apple/callback", h.AppleCodeExchange)
	router.POST("/login", h.EmailLogin)
	router.POST("/signup", h.EmailSignUp)
	router.GET("/refresh-token", h.RefreshToken)
	router.GET("/email-verify", h.EmailVerify)
	router.POST("/password-reset", h.PasswordReset)
	router.POST("/password-reset/confirm", h.PasswordResetConfirm)
	router.POST("/password-token-check", h.PasswordResetTokenCheck)
	router.POST("/password-update", h.UpdatePassword, middlewares.AuthMiddleware("USER", "ADMIN"))

	router.GET("/logout", h.Logout, middlewares.AuthMiddleware("USER", "ADMIN"))
	router.POST("/resend-verification-email", h.ResendEmailVerification)
}
