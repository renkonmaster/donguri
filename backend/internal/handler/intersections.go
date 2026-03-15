package handler

import (
	"context"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

// GetRoomIntersections implements [api.Handler].
func (h *Handler) GetRoomIntersections(ctx context.Context, params api.GetRoomIntersectionsParams) (*api.RoomIntersectionsResponse, error) {
	pairs, err := h.repo.GetIntersectingEdgePairs(ctx, params.RoomID, nil)
	if err != nil {
		return nil, err
	}

	return &api.RoomIntersectionsResponse{Intersections: toAPIIntersectingEdgePairs(pairs)}, nil
}

func toAPIIntersectingEdgePairs(pairs []repository.IntersectingEdgePair) []api.IntersectingEdgePair {
	intersections := make([]api.IntersectingEdgePair, 0, len(pairs))
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
		})
	}

	return intersections
}
