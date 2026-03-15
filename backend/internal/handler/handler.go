package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"
)

type Handler struct {
	api.UnimplementedHandler
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

// CreateMessage implements [api.Handler].
func (h *Handler) CreateMessage(ctx context.Context, req *api.CreateMessageRequest, params api.CreateMessageParams) (*api.Message, error) {
	if req == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				Message: "request body is required",
			},
		}
	}

	entity, err := h.repo.CreateMessage(ctx, repository.CreateMessageParams{
		RoomID:     params.RoomID,
		SenderID:   params.XPlayerID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		return nil, err
	}

	message := &api.Message{
		ID:         entity.ID,
		RoomID:     entity.RoomID,
		SenderID:   entity.SenderID,
		ReceiverID: req.ReceiverID,
		Content:    entity.Content,
		CreatedAt:  entity.CreatedAt,
	}

	if payload, marshalErr := json.Marshal(message); marshalErr == nil {
		h.publishRoomEvent(params.RoomID, "message_received", payload)
	}

	return message, nil
}

// CreateSwapIntent implements [api.Handler].
func (h *Handler) CreateSwapIntent(ctx context.Context, req *api.SwapIntentRequest, params api.CreateSwapIntentParams) (*api.SwapIntentResponse, error) {
	if req == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				Message: "request body is required",
			},
		}
	}

	if err := h.repo.SetSwapIntent(ctx, repository.SetSwapIntentParams{
		RoomID:     params.RoomID,
		SenderID:   params.XPlayerID,
		ReceiverID: req.TargetPlayerID,
		NeedsSwap:  true,
	}); err != nil {
		return nil, err
	}

	response := &api.SwapIntentResponse{
		Matched:    false,
		RoomStatus: api.RoomStatusPlaying,
	}

	payload := map[string]any{
		"sender_id":        params.XPlayerID,
		"target_player_id": req.TargetPlayerID,
		"matched":          response.Matched,
		"room_status":      response.RoomStatus,
	}
	if b, marshalErr := json.Marshal(payload); marshalErr == nil {
		h.publishRoomEvent(params.RoomID, "connection_updated", b)
	}

	return response, nil
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
	return api.SubscribeRoomStreamOK{
		Data: h.newRoomStreamReader(ctx, params.RoomID.String()),
	}, nil
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
