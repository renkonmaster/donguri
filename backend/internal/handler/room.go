package handler

import (
	"context"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
)

// GetRoom implements [api.Handler].
func (h *Handler) GetRoom(ctx context.Context, params api.GetRoomParams) (*api.RoomStateResponse, error) {
	state, err := h.repo.GetRoomState(ctx, params.RoomID, params.XPlayerID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusNotFound,
			Response: api.Error{
				Message: "room not found",
			},
		}
	}

	players := make([]api.Player, 0, len(state.Players))
	for _, p := range state.Players {
		players = append(players, api.Player{
			ID:         p.ID,
			Name:       p.Name,
			OrderIndex: p.OrderIndex,
			Location: api.Location{
				Lat: p.Lat,
				Lng: p.Lng,
			},
		})
	}

	var timeLeft api.OptNilInt
	if state.TimeLeftSeconds == nil {
		timeLeft.SetToNull()
	} else {
		timeLeft.SetTo(*state.TimeLeftSeconds)
	}

	return &api.RoomStateResponse{
		Status:            api.RoomStatus(state.Status),
		TimeLeft:          timeLeft,
		IntersectionCount: state.IntersectionCount,
		Players:           players,
		MyIntents:         state.MyIntents,
		ReceivedIntents:   state.ReceivedIntents,
	}, nil
}
