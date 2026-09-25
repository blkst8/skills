package middlewares

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTAuthentication returns a middleware that validates Bearer tokens signed
// with HS256 using the given secret. On success the parsed claims are stored
// in the request context under the "user" key.
func JWTAuthentication(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				return echo.ErrUnauthorized
			}

			claims := jwt.RegisteredClaims{}
			parsed, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			if err != nil || !parsed.Valid {
				return echo.ErrUnauthorized
			}

			c.Set("user", claims)

			return next(c)
		}
	}
}
