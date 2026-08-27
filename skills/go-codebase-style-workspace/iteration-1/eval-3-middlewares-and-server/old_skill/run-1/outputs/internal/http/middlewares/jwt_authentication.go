// Package middlewares provides HTTP middleware for the Echo server.
//
// Every middleware is a constructor returning an echo.MiddlewareFunc so it
// can be registered globally via e.Use() or scoped to a route group via
// group.Use() in internal/http/server.go.
package middlewares

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"yourproject/internal/config"
)

// userClaimsKey is the Echo context key under which the validated JWT
// claims are stored for downstream handlers.
const userClaimsKey = "user"

// bearerPrefix is the expected scheme of the Authorization header.
const bearerPrefix = "Bearer "

// Claims represents the JWT payload issued by the authentication service.
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// JWTAuthentication returns an Echo middleware that authenticates requests
// with a JWT presented in the Authorization header.
//
// It expects a "Bearer <token>" header, verifies the HMAC signature against
// config.C.HTTPServer.JWTSecret, and stores the parsed claims in the Echo
// context under the "user" key. It returns echo.ErrUnauthorized when the
// header is missing, malformed, or the token fails validation.
func JWTAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authHeader := ctx.Request().Header.Get(echo.HeaderAuthorization)

			tokenString, found := strings.CutPrefix(authHeader, bearerPrefix)
			if !found || tokenString == "" {
				return echo.ErrUnauthorized
			}

			claims, err := validateToken(tokenString)
			if err != nil {
				return echo.ErrUnauthorized
			}

			ctx.Set(userClaimsKey, claims)

			return next(ctx)
		}
	}
}

// ClaimsFromContext returns the JWT claims stored by JWTAuthentication.
//
// Handlers on protected routes use it to read the authenticated user.
func ClaimsFromContext(ctx echo.Context) (*Claims, error) {
	claims, ok := ctx.Get(userClaimsKey).(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("no JWT claims found in request context")
	}

	return claims, nil
}

// validateToken parses and verifies a signed JWT and returns its claims.
func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(config.C.HTTPServer.JWTSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}
