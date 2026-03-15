package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gorm.io/gorm"
)

type RoomPlayerState struct {
	ID         uuid.UUID
	Name       string
	OrderIndex int
	Lat        float64
	Lng        float64
}

type RoomState struct {
	Status            string
	TimeLeftSeconds   *int
	IntersectionCount int
	Players           []RoomPlayerState
	MyIntents         []uuid.UUID
	ReceivedIntents   []uuid.UUID
}

// GetRoomState returns room status and player-related state for GET /api/rooms/{room_id}.
func (r *Repository) GetRoomState(ctx context.Context, roomID, playerID uuid.UUID) (*RoomState, error) {
	var room database.RoomEntity
	if err := r.db.WithContext(ctx).
		Select("status", "expires_at").
		Take(&room, "id = ?", roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("select room: %w", err)
	}

	intersectionCount, err := r.GetRoomIntersectionCount(ctx, roomID, nil)
	if err != nil {
		return nil, err
	}

	players, err := r.selectRoomPlayers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	myIntents, err := r.selectIntentIDs(ctx, roomID, "sender_id", playerID, "receiver_id")
	if err != nil {
		return nil, err
	}

	receivedIntents, err := r.selectIntentIDs(ctx, roomID, "receiver_id", playerID, "sender_id")
	if err != nil {
		return nil, err
	}

	var timeLeftSeconds *int
	if room.ExpiresAt != nil {
		seconds := int(time.Until(*room.ExpiresAt).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		timeLeftSeconds = &seconds
	}

	return &RoomState{
		Status:            room.Status,
		TimeLeftSeconds:   timeLeftSeconds,
		IntersectionCount: intersectionCount,
		Players:           players,
		MyIntents:         myIntents,
		ReceivedIntents:   receivedIntents,
	}, nil
}

func (r *Repository) selectRoomPlayers(ctx context.Context, roomID uuid.UUID) ([]RoomPlayerState, error) {
	if r.db.Name() == "sqlite" {
		var rows []struct {
			ID         uuid.UUID `gorm:"column:id"`
			Name       string    `gorm:"column:name"`
			OrderIndex int       `gorm:"column:order_index"`
			Location   string    `gorm:"column:location"`
		}
		if err := r.db.WithContext(ctx).
			Table("players").
			Select("id", "name", "order_index", "location").
			Where("room_id = ?", roomID).
			Order("order_index ASC").
			Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("select players: %w", err)
		}

		players := make([]RoomPlayerState, 0, len(rows))
		for _, row := range rows {
			point, err := parseEWKTPoint(row.Location)
			if err != nil {
				return nil, fmt.Errorf("parse player location: %w", err)
			}
			players = append(players, RoomPlayerState{
				ID:         row.ID,
				Name:       row.Name,
				OrderIndex: row.OrderIndex,
				Lat:        point.Lat,
				Lng:        point.Lng,
			})
		}

		return players, nil
	}

	var rows []struct {
		ID         uuid.UUID `gorm:"column:id"`
		Name       string    `gorm:"column:name"`
		OrderIndex int       `gorm:"column:order_index"`
		Lat        float64   `gorm:"column:lat"`
		Lng        float64   `gorm:"column:lng"`
	}
	if err := r.db.WithContext(ctx).
		Table("players").
		Select("id", "name", "order_index", "ST_Y(location::geometry) AS lat", "ST_X(location::geometry) AS lng").
		Where("room_id = ?", roomID).
		Order("order_index ASC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("select players: %w", err)
	}

	players := make([]RoomPlayerState, 0, len(rows))
	for _, row := range rows {
		players = append(players, RoomPlayerState{
			ID:         row.ID,
			Name:       row.Name,
			OrderIndex: row.OrderIndex,
			Lat:        row.Lat,
			Lng:        row.Lng,
		})
	}

	return players, nil
}

func (r *Repository) selectIntentIDs(
	ctx context.Context,
	roomID uuid.UUID,
	filterColumn string,
	filterValue uuid.UUID,
	targetColumn string,
) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("connections").
		Where("room_id = ?", roomID).
		Where(filterColumn+" = ?", filterValue).
		Where("needs_swap = TRUE").
		Order(targetColumn+" ASC").
		Pluck(targetColumn, &ids).Error
	if err != nil {
		return nil, fmt.Errorf("select intents (%s): %w", targetColumn, err)
	}

	return ids, nil
}

type latLng struct {
	Lat float64
	Lng float64
}

// parseEWKTPoint parses "SRID=4326;POINT(lng lat)" returned by SQLite for location columns.
func parseEWKTPoint(ewkt string) (latLng, error) {
	const prefix = "POINT("
	idx := strings.Index(ewkt, prefix)
	if idx == -1 {
		return latLng{}, fmt.Errorf("unexpected EWKT format: %q", ewkt)
	}
	body := strings.TrimSuffix(ewkt[idx+len(prefix):], ")")
	var lng, lat float64
	if _, err := fmt.Sscanf(body, "%f %f", &lng, &lat); err != nil {
		return latLng{}, fmt.Errorf("parse EWKT coordinates %q: %w", ewkt, err)
	}

	return latLng{Lat: lat, Lng: lng}, nil
}
