package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ras0q/go-backend-template/internal/api"
	"github.com/ras0q/go-backend-template/internal/repository"
	"github.com/ras0q/go-backend-template/internal/service/photo"
)

type Handler struct {
	photo *photo.Service
	repo  *repository.Repository
}

// CreateMessage implements [api.Handler].
func (h *Handler) CreateMessage(ctx context.Context, req *api.CreateMessageRequest, params api.CreateMessageParams) (*api.Message, error) {
	panic("unimplemented")
}

// CreateSwapIntent implements [api.Handler].
func (h *Handler) CreateSwapIntent(ctx context.Context, req *api.SwapIntentRequest, params api.CreateSwapIntentParams) (*api.SwapIntentResponse, error) {
	panic("unimplemented")
}

// DeleteMyDirectionalIntent implements [api.Handler].
func (h *Handler) DeleteMyDirectionalIntent(ctx context.Context, req *api.DirectionalIntentRequest, params api.DeleteMyDirectionalIntentParams) (*api.DirectionalIntentResponse, error) {
	panic("unimplemented")
}

// DeleteSwapIntent implements [api.Handler].
func (h *Handler) DeleteSwapIntent(ctx context.Context, params api.DeleteSwapIntentParams) (*api.DeleteSwapIntentResponse, error) {
	panic("unimplemented")
}

// GetMessages implements [api.Handler].
func (h *Handler) GetMessages(ctx context.Context, params api.GetMessagesParams) ([]api.Message, error) {
	panic("unimplemented")
}

// GetRoom implements [api.Handler].
func (h *Handler) GetRoom(ctx context.Context, params api.GetRoomParams) (*api.RoomStateResponse, error) {
	panic("unimplemented")
}

// JoinRoom implements [api.Handler].
func (h *Handler) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.JoinRoomResponse, error) {
	panic("unimplemented")
}

// PatchMyDirectionalIntent implements [api.Handler].
func (h *Handler) PatchMyDirectionalIntent(ctx context.Context, req *api.DirectionalIntentRequest, params api.PatchMyDirectionalIntentParams) (*api.DirectionalIntentResponse, error) {
	panic("unimplemented")
}

// SubscribeRoomStream implements [api.Handler].
func (h *Handler) SubscribeRoomStream(ctx context.Context, params api.SubscribeRoomStreamParams) (api.SubscribeRoomStreamOK, error) {
	panic("unimplemented")
}

func New(
	photo *photo.Service,
	repo *repository.Repository,
) *Handler {
	return &Handler{
		photo,
		repo,
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
