package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/db"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
)

func AuthMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errorhandler.ErrorUnauthorized("Authorization header is required")
			}
			Token := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := utils.VerifyJWT(Token)
			if err != nil {
				return errorhandler.ErrorUnauthorized("Invalid access token")
			}

			sid, err := uuid.Parse(claims.Sid)
			if err != nil {
				return errorhandler.ErrorUnauthorized("invalid session ID")
			}
			if claims.Role == string(types.RoleTypeAdminMode) {
				if c.Request().Method != http.MethodGet {
					return errorhandler.ErrorForbidden("admin_mode can only view resources")
				}
			}

			if claims.Role != string(types.RoleTypeAdminMode) {
				// TODO: implement cache
				sess, err := db.SqlcQuery.GetSessionByID(c.Request().Context(), sid)
				if err != nil {
					return errorhandler.ErrorUnauthorized("session expired")
				}

				if time.Now().After(sess.ExpiresAt) {
					return errorhandler.ErrorUnauthorized("session expired")
				}

				if sess.RevokedAt != nil && time.Now().After(*sess.RevokedAt) {
					return errorhandler.ErrorUnauthorized("You have already signed out of this session.")
				}

			}
			// TODO: optimze using queue
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = db.SqlcQuery.UpdateSessionLastSeen(ctx, sid)
			}()

			userID := claims.Subject
			if userID == "" {
				return errorhandler.ErrorUnauthorized("Missing user ID in token claims")
			}

			c.Set("claims", claims)

			if !isRoleAllowed(claims.Role, allowedRoles) {
				return errorhandler.ErrorForbidden("insufficient permissions")
			}
			return next(c)
		}
	}
}

func ApiKeyAuthMiddleware(allowedPaths ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqPath := c.Request().URL.Path

			// check allowed paths
			for _, pattern := range allowedPaths {
				if strings.HasSuffix(pattern, "*") {
					// prefix match
					prefix := strings.TrimSuffix(pattern, "*")
					if strings.HasPrefix(reqPath, prefix) {
						return next(c)
					}
				} else {
					// exact match
					if reqPath == pattern {
						return next(c)
					}
				}
			}

			apiKey := c.Request().Header.Get("X-Api-Key")
			if apiKey == "" {
				return errorhandler.ErrorUnauthorized("X-Api-key header is required")
			}

			if apiKey != config.Cfg.APIKey {
				return errorhandler.ErrorUnauthorized("Invalid X-Api-key")
			}
			return next(c)
		}
	}
}

func isRoleAllowed(userRole string, allowedRoles []string) bool {
	if userRole == string(types.RoleTypeAdminMode) {
		return true
	}
	return slices.Contains(allowedRoles, userRole)
}
