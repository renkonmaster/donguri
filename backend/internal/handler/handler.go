package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"
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

// DeleteMyDirectionalIntent implements [api.Handler].
func (h *Handler) DeleteMyDirectionalIntent(_ context.Context, _ *api.DirectionalIntentRequest, _ api.DeleteMyDirectionalIntentParams) (*api.DirectionalIntentResponse, error) {
	panic("unimplemented")
}

// DeleteSwapIntent implements [api.Handler].
func (h *Handler) DeleteSwapIntent(_ context.Context, _ api.DeleteSwapIntentParams) (*api.DeleteSwapIntentResponse, error) {
	panic("unimplemented")
}

// PatchMyDirectionalIntent implements [api.Handler].
func (h *Handler) PatchMyDirectionalIntent(_ context.Context, _ *api.DirectionalIntentRequest, _ api.PatchMyDirectionalIntentParams) (*api.DirectionalIntentResponse, error) {
	panic("unimplemented")
}

// SubscribeRoomStream implements [api.Handler].
func (h *Handler) SubscribeRoomStream(ctx context.Context, params api.SubscribeRoomStreamParams) (api.SubscribeRoomStreamOK, error) {
	return api.SubscribeRoomStreamOK{
		Data: h.newRoomStreamReader(ctx, params.RoomID.String()),
	}, nil
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	if statusErr := messageValidationError(err); statusErr != nil {
		return statusErr
	}

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
