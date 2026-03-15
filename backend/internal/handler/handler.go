package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ras0q/go-backend-template/internal/api"
	"github.com/ras0q/go-backend-template/internal/repository"
	"github.com/ras0q/go-backend-template/internal/service/stream"
)

type Handler struct {
	repo *repository.Repository
	hub  *stream.Hub
}

func New(
	repo *repository.Repository,
	hub *stream.Hub,
) *Handler {
	return &Handler{
		repo: repo,
		hub:  hub,
	}
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
