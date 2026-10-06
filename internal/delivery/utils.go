package delivery

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Neratus/geoguide/internal/api"
	"github.com/Neratus/geoguide/internal/delivery/middleware"
	"github.com/Neratus/geoguide/internal/domain"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefStringSlice(s *[]string) []string {
	if s == nil {
		return []string{}
	}
	return *s
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func parseCoordinatesString(coordStr string) *api.Coordinates {
	if coordStr == "" {
		return nil
	}
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return nil
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil
	}
	return &api.Coordinates{
		Lat: lat,
		Lng: lng,
	}
}

func convertCoordinatesToAPI(coord domain.Coordinates) *api.Coordinates {
	return &api.Coordinates{
		Lat: coord.Lat(),
		Lng: coord.Lng(),
	}
}

func float32Ptr(f float64) *float32 {
	v := float32(f)
	return &v
}

func mapOptionalDate(d *openapi_types.Date) *time.Time {
	if d == nil {
		return nil
	}
	return &d.Time
}

func mapOptionalFloat32(f *float32) *float64 {
	if f == nil {
		return nil
	}
	v := float64(*f)
	return &v
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func respondError(ctx echo.Context, code int, message string) error {
	return ctx.JSON(code, api.ErrorResponse{
		Error: message,
	})
}

func getUserID(ctx echo.Context) (domain.UserID, error) {
	val := ctx.Request().Context().Value(middleware.UserIDKey)
	if val == nil {
		return domain.UserID{}, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing user ID in context")
	}
	userID, ok := val.(domain.UserID)
	if !ok {
		return domain.UserID{}, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: invalid user ID type")
	}
	return userID, nil
}

func getIsModerator(ctx echo.Context) bool {
	val := ctx.Request().Context().Value(middleware.IsModeratorKey)
	if val == nil {
		return false
	}
	isMod, ok := val.(bool)
	return ok && isMod
}

func datePtr(t *openapi_types.Date) *openapi_types.Date { return t }
func timePtr(t time.Time) *time.Time                    { return &t }

func derefIntWithDefault(val *int, defaultVal int) int {
	if val != nil {
		return *val
	}
	return defaultVal
}

func derefStringWithDefault(val *string, defaultVal string) string {
	if val != nil {
		return *val
	}
	return defaultVal
}

func derefBoolWithDefault(val *bool, defaultVal bool) bool {
	if val != nil {
		return *val
	}
	return defaultVal
}

func parseDate(s string) *openapi_types.Date {
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d := openapi_types.Date{Time: t}
		return &d
	}
	return nil
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

func splitBearer(authHeader string) []string {
	parts := splitSpace(authHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}
	return parts
}

func splitSpace(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func ptrEmail(e openapi_types.Email) *openapi_types.Email { return &e }
