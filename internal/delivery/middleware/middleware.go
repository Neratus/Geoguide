package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	IsModeratorKey contextKey = "isModerator"
	RoleKey        contextKey = "role"
	RawTokenKey    contextKey = "rawToken"
)

func JWTAuthEcho(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			publicGetPrefixes := []string{
				"/api/v1/countries", "/api/v1/cities", "/api/v1/places",
				"/api/v1/files", "/api/v1/holidays",
				"/api/v2/countries", "/api/v2/cities", "/api/v2/places",
				"/api/v2/files", "/api/v2/holidays",
			}
			publicPaths := []string{
				"/api/v1/auth/register", "/api/v1/auth/login",
				"/api/v1/auth/2fa/verify", "/api/v1/auth/verify",
				"/api/v2/auth/register", "/api/v2/auth/login",
				"/api/v2/auth/2fa/verify", "/api/v2/auth/verify",
				"/openapi.yaml", "/swagger",
				"/api/v2/graphql", "/api/v2/playground",
			}

			isPublic := false
			if req.Method == http.MethodGet {
				for _, prefix := range publicGetPrefixes {
					if strings.HasPrefix(req.URL.Path, prefix) {
						isPublic = true
						break
					}
				}
			}
			for _, p := range publicPaths {
				if strings.HasPrefix(req.URL.Path, p) {
					isPublic = true
					break
				}
			}

			authHeader := req.Header.Get("Authorization")

			if authHeader == "" {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			tokenType, _ := claims["type"].(string)
			if tokenType == "refresh" {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "refresh token cannot be used for API access")
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "missing user_id in token")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				if isPublic {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid user_id format")
			}

			role, _ := claims["role"].(string)
			isModerator := (role == domain.UserRoleModerator || role == domain.UserRoleAnalyst)

			ctx := context.WithValue(req.Context(), UserIDKey, domain.UserID(userID))
			ctx = context.WithValue(ctx, IsModeratorKey, isModerator)
			ctx = context.WithValue(ctx, RoleKey, role)
			ctx = context.WithValue(ctx, RawTokenKey, tokenStr)

			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func LoggerEcho(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		status := c.Response().Status
		if status == 0 {
			status = http.StatusOK
		}

		slog.Info("request",
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", c.Request().RemoteAddr,
		)

		return err
	}
}

func CORSEcho(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}

		return next(c)
	}
}
