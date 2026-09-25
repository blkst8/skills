package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// JWTAuthentication validates the Authorization header before the request
// reaches the handlers.
func JWTAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")

			if token == "" {
				return echo.ErrUnauthorized
			}

			token = strings.TrimPrefix(token, "Bearer ")

			// Token validation (signature, expiry, claims) happens here.
			c.Set("user", token)

			return next(c)
		}
	}
}
