package delivery

import (
	"net/http"

	"github.com/Neratus/geoguide/internal/api"
	"github.com/Neratus/geoguide/internal/delivery/middleware"
	"github.com/labstack/echo/v4"
)

func NewRouter(srv *Server, jwtSecret []byte) *echo.Echo {
	e := echo.New()

	e.Use(middleware.LoggerEcho)
	e.Use(middleware.CORSEcho)
	e.Use(middleware.JWTAuthEcho(jwtSecret))

	api.RegisterHandlers(e, srv)

	e.POST("/api/v1/auth/avatar", srv.UploadAvatar)

	e.File("/openapi.yaml", "api/openapi.yaml")
	e.GET("/swagger", func(c echo.Context) error {
		html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>GeoGuide API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '/openapi.yaml',
                dom_id: '#swagger-ui',
                persistAuthorization: true,
                deepLinking: true,
            });
        };
    </script>
</body>
</html>`
		return c.HTML(http.StatusOK, html)
	})

	return e
}
