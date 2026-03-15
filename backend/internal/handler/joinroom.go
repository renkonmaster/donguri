package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

// JoinRoom implements [api.Handler].
func (h *Handler) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.JoinRoomResponse, error) {
	const maxPlayersPerRoom = 4

	var (
		roomID   uuid.UUID
		playerID uuid.UUID
		joinedCount int
	)

	err := h.repo.Transaction(ctx, func(repo *repository.Repository) error {
		room, err := repo.FindMatchingRoom(ctx, maxPlayersPerRoom)
		if err != nil {
			return err
		}
		if room == nil {
			room, err = repo.CreateRoom(ctx)
			if err != nil {
				return err
			}
		}

		count, err := repo.CountPlayersInRoom(ctx, room.ID)
		if err != nil {
			return err
		}

		player, err := repo.CreatePlayer(ctx, room.ID, req.GetName(), req.GetLat(), req.GetLng(), count)
		if err != nil {
			return err
		}

		roomID = room.ID
		playerID = player.ID
		joinedCount = count + 1
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("join room: %w", err)
	}

	if payload, marshalErr := json.Marshal(map[string]any{
		"room_id":      roomID,
		"player_id":    playerID,
		"joined_count": joinedCount,
		"event":        "joined",
	}); marshalErr == nil {
		h.publishRoomEvent(roomID, "room_updated", payload)
	}

	return &api.JoinRoomResponse{
		RoomID:   roomID,
		PlayerID: playerID,
	}, nil
}
