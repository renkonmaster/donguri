package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

type Handler struct {
	api.UnimplementedHandler
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	if apiErr, ok := err.(*api.ErrorStatusCode); ok {
		return apiErr
	}

	slog.ErrorContext(ctx, "internal server error", "error", err)

	return &api.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: api.Error{
			Message: "internal server error",
		},
	}
}
