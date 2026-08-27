package middlewares

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/blkst8/invoice-service/internal/config"
)

// JWTAuthentication validates the Bearer JWT on protected routes using the
// HMAC secret from config. Valid claims are stored under the "user" context
// key.
func JWTAuthentication() echo.MiddlewareFunc {
	secret := []byte(config.C.JWT.Secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authHeader := ctx.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.ErrUnauthorized
			}

			token, err := jwt.Parse(strings.TrimPrefix(authHeader, "Bearer "), func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				return echo.ErrUnauthorized
			}

			ctx.Set("user", token.Claims)
			return next(ctx)
		}
	}
}
