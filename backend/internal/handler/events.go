package handler

import (
	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/service/stream"
)

func (h *Handler) publishRoomEvent(roomID uuid.UUID, eventName string, data []byte) {
	h.hub.Publish(roomID.String(), stream.FormatEvent(stream.Event{
		Name: eventName,
		Data: data,
	}))
}
