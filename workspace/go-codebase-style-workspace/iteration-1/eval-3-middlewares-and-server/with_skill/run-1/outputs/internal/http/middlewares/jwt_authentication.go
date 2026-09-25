package middlewares

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTAuthentication returns a middleware that requires a valid Bearer JWT
// signed with the given HMAC secret. On success the parsed claims are stored
// in the request context under the "user" key.
func JWTAuthentication(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")

			if token == "" {
				return echo.ErrUnauthorized
			}

			claims, err := validateToken(secret, strings.TrimPrefix(token, "Bearer "))
			if err != nil {
				return echo.ErrUnauthorized
			}

			c.Set("user", claims)

			return next(c)
		}
	}
}

// validateToken parses and validates an HMAC-signed JWT and returns its claims.
func validateToken(secret string, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
