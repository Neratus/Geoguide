package middleware

import (
	"context"
	"fmt"
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
)

func JWTAuthEcho(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			if req.Method == http.MethodGet {
				publicGetPrefixes := []string{
					"/api/v1/countries",
					"/api/v1/cities",
					"/api/v1/places",
					"/api/v1/files",
					"/api/v1/holidays",
				}
				for _, prefix := range publicGetPrefixes {
					fmt.Println(prefix, req.URL.Path, strings.HasPrefix(req.URL.Path, prefix))
					if strings.HasPrefix(req.URL.Path, prefix) {
						return next(c)
					}
				}
			}

			publicPaths := []string{
				"/api/v1/auth/register",
				"/api/v1/auth/login",
				"/api/v1/auth/2fa/verify",
				"/api/v1/auth/verify",
				"/openapi.yaml",
				"/swagger",
			}
			for _, p := range publicPaths {
				fmt.Println(p, req.URL.Path, strings.HasPrefix(req.URL.Path, p))
				if strings.HasPrefix(req.URL.Path, p) {
					return next(c)
				}
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
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
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing user_id in token")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid user_id format")
			}

			role, _ := claims["role"].(string)
			isModerator := (role == domain.UserRoleModerator || role == domain.UserRoleAnalyst)

			ctx := context.WithValue(req.Context(), UserIDKey, domain.UserID(userID))
			ctx = context.WithValue(ctx, IsModeratorKey, isModerator)
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
