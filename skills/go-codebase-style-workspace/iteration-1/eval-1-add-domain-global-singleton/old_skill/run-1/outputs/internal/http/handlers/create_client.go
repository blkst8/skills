package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"yourproject/internal/app"
	"yourproject/internal/log"
	"yourproject/internal/repository"
	"yourproject/internal/usecase"
)

// CreateClientRequest is the request body for creating a client.
type CreateClientRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateClientResponse is the response body for a created client.
type CreateClientResponse struct {
	ID        uint32    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateClient handles the creation of a new client.
//
// It binds the request, delegates to the client usecase (which validates the
// payload, inserts the client and re-fetches the stored record) and returns
// the stored client.
//
// Request body:
//   - name (string, required): display name of the client
//   - email (string, required): unique email address of the client
//
// Returns HTTP 200 on success, 400 on validation error, 409 when the email
// already exists, 500 on server error.
func CreateClient(ctx echo.Context) error {
	var request CreateClientRequest
	if err := ctx.Bind(&request); err != nil {
		log.Logger.Error("failed to bind request", zap.Error(err))
		return err
	}

	created, err := app.A.Service.Client.Create(ctx.Request().Context(), usecase.CreateClientInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		log.Logger.Error("failed to create client", zap.Error(err))

		switch {
		case errors.Is(err, usecase.ErrInvalidClientInput):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrClientAlreadyExists):
			return echo.NewHTTPError(http.StatusConflict, "client already exists")
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, CreateClientResponse{
		ID:        created.ID,
		Name:      created.Name,
		Email:     created.Email,
		CreatedAt: created.CreatedAt,
	})
}
