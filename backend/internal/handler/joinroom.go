package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
)

const (
	maxPlayersPerRoom = 8
	gameDuration      = 5 * time.Minute
	dbOpTimeout       = 5 * time.Second
)

// gameStartDelay is a variable so that tests can shorten the delay and
// deterministically cover the "full → delay → room_started" path.
var gameStartDelay = 5 * time.Second

// JoinRoom implements [api.Handler].
func (h *Handler) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.JoinRoomResponse, error) {
	if req == nil {
		return nil, &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				Message: "request is invalied",
			},
		}
	}

	var (
		roomID      uuid.UUID
		playerID    uuid.UUID
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

	if joinedCount == maxPlayersPerRoom {
		go h.scheduleGameStart(roomID)
	}

	return &api.JoinRoomResponse{
		RoomID:   roomID,
		PlayerID: playerID,
	}, nil
}

func (h *Handler) scheduleGameStart(roomID uuid.UUID) {
	time.Sleep(gameStartDelay)

	ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
	defer cancel()

	changed, err := h.repo.MarkRoomPlayingIfFull(ctx, roomID, maxPlayersPerRoom, maxPlayersPerRoom, gameDuration)
	if err != nil {
		slog.Error("scheduleGameStart: failed to mark room playing", "room_id", roomID, "error", err)

		return
	}
	if !changed {
		return
	}

	if err := h.repo.ShufflePlayerOrderIndices(ctx, roomID); err != nil {
		slog.Error("scheduleGameStart: failed to shuffle player order indices", "room_id", roomID, "error", err)
	}

	if payload, marshalErr := json.Marshal(map[string]any{"room_id": roomID}); marshalErr == nil {
		h.publishRoomEvent(roomID, "room_started", payload)
	}

	go h.scheduleGameTimeout(roomID, gameDuration)
}

func (h *Handler) scheduleGameTimeout(roomID uuid.UUID, duration time.Duration) {
	time.Sleep(duration)

	ctx, cancel := context.WithTimeout(context.Background(), dbOpTimeout)
	defer cancel()

	changed, err := h.repo.ExpireRoomIfPlaying(ctx, roomID)
	if err != nil {
		slog.Error("scheduleGameTimeout: failed to expire room", "room_id", roomID, "error", err)

		return
	}
	if !changed {
		return
	}

	if payload, marshalErr := json.Marshal(map[string]any{"room_id": roomID, "reason": "timeout"}); marshalErr == nil {
		h.publishRoomEvent(roomID, "room_updated", payload)
	}
}
