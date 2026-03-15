package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

// PutConnection implements [api.Handler].
func (h *Handler) PutConnection(ctx context.Context, req *api.ConnectionRequest, params api.PutConnectionParams) (*api.ConnectionResponse, error) {
	if req == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				Message: "request body is required",
			},
		}
	}

	result, err := h.repo.SetSwapIntent(ctx, repository.SetSwapIntentParams{
		RoomID:     params.RoomID,
		SenderID:   params.XPlayerID,
		ReceiverID: params.TargetID,
		NeedsSwap:  req.NeedsSwap,
	})
	if err != nil {
		return nil, err
	}

	roomStatus := api.RoomStatus(result.RoomStatus)

	response := &api.ConnectionResponse{
		Matched:    result.Matched,
		RoomStatus: roomStatus,
	}

	payload := map[string]any{
		"sender_id":        params.XPlayerID,
		"target_player_id": params.TargetID,
		"needs_swap":       req.NeedsSwap,
		"matched":          response.Matched,
		"room_status":      response.RoomStatus,
	}
	if response.Matched {
		intersections := []api.IntersectingEdgePair{}
		pairs, pairErr := h.repo.GetIntersectingEdgePairs(ctx, params.RoomID, nil)
		if pairErr == nil {
			intersections = make([]api.IntersectingEdgePair, 0, len(pairs))
			for _, pair := range pairs {
				intersections = append(intersections, api.IntersectingEdgePair{
					First: api.OrderedEdge{
						StartOrderIndex: pair.First.StartOrderIndex,
						EndOrderIndex:   pair.First.EndOrderIndex,
					},
					Second: api.OrderedEdge{
						StartOrderIndex: pair.Second.StartOrderIndex,
						EndOrderIndex:   pair.Second.EndOrderIndex,
					},
					IntersectionLocation: api.Location{
						Lat: pair.Location.Lat,
						Lng: pair.Location.Lng,
					},
				})
			}
		}

		payload["intersections"] = intersections
	}
	if b, marshalErr := json.Marshal(payload); marshalErr == nil {
		h.publishRoomEvent(params.RoomID, "room_updated", b)
	}

	return response, nil
}
