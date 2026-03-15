package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

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
	if entity.ReceiverID == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusInternalServerError,
			Response:   api.Error{Message: "receiver_id is missing in stored message"},
		}
	}

	message := &api.Message{
		ID:         entity.ID,
		RoomID:     entity.RoomID,
		SenderID:   entity.SenderID,
		ReceiverID: *entity.ReceiverID,
		Content:    entity.Content,
		CreatedAt:  entity.CreatedAt,
	}

	if payload, marshalErr := json.Marshal(message); marshalErr == nil {
		h.publishRoomEvent(params.RoomID, "message_received", payload)
	}

	return message, nil
}

// GetMessages implements [api.Handler].
func (h *Handler) GetMessages(ctx context.Context, params api.GetMessagesParams) ([]api.Message, error) {
	entities, err := h.repo.GetMessagesRelatedToPlayer(ctx, params.RoomID, params.XPlayerID)
	if err != nil {
		return nil, err
	}

	messages := make([]api.Message, 0, len(entities))
	for _, entity := range entities {
		if entity.ReceiverID == nil {
			continue
		}

		messages = append(messages, api.Message{
			ID:         entity.ID,
			RoomID:     entity.RoomID,
			SenderID:   entity.SenderID,
			ReceiverID: *entity.ReceiverID,
			Content:    entity.Content,
			CreatedAt:  entity.CreatedAt,
		})
	}

	return messages, nil
}

func messageValidationError(err error) *api.ErrorStatusCode {
	if errors.Is(err, repository.ErrReceiverIDRequired) ||
		errors.Is(err, repository.ErrSenderReceiverSame) ||
		errors.Is(err, repository.ErrPlayerNotFoundInRoom) ||
		errors.Is(err, repository.ErrPlayersNotAdjacent) {
		return &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				Message: err.Error(),
			},
		}
	}

	return nil
}
