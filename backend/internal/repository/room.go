package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gorm.io/gorm"
)

func (r *Repository) FindMatchingRoom(ctx context.Context, maxPlayers int) (*database.RoomEntity, error) {
	var room database.RoomEntity
	err := r.db.WithContext(ctx).
		Where("status = ?", database.RoomStatusMatching).
		Where("(SELECT COUNT(*) FROM players WHERE room_id = rooms.id) < ?", maxPlayers).
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find matching room: %w", err)
	}

	return &room, nil
}

func (r *Repository) CreateRoom(ctx context.Context) (*database.RoomEntity, error) {
	var room database.RoomEntity
	room.ID = uuid.New()
	room.Status = database.RoomStatusMatching
	if err := r.db.WithContext(ctx).Create(&room).Error; err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	return &room, nil
}

func (r *Repository) CountPlayersInRoom(ctx context.Context, roomID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(database.PlayerEntity)).
		Where("room_id = ?", roomID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count players in room: %w", err)
	}

	return int(count), nil
}

func (r *Repository) MarkRoomPlayingIfFull(
	ctx context.Context,
	roomID uuid.UUID,
	joinedCount int,
	maxPlayers int,
	duration time.Duration,
) (changed bool, err error) {
	if joinedCount < maxPlayers {
		return false, nil
	}

	now := time.Now().UTC()
	expiresAt := now.Add(duration)

	tx := r.db.WithContext(ctx).
		Model(new(database.RoomEntity)).
		Where("id = ? AND status = ?", roomID, database.RoomStatusMatching).
		Updates(map[string]any{
			"status":     database.RoomStatusPlaying,
			"start_at":   now,
			"expires_at": expiresAt,
		})
	if tx.Error != nil {
		return false, fmt.Errorf("mark room playing: %w", tx.Error)
	}

	return tx.RowsAffected > 0, nil
}

func (r *Repository) ExpireRoomIfPlaying(ctx context.Context, roomID uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(new(database.RoomEntity)).
		Where("id = ? AND status = ? AND expires_at <= ?", roomID, database.RoomStatusPlaying, now).
		Updates(map[string]any{
			"status":     database.RoomStatusFinished,
			"updated_at": now,
		})
	if tx.Error != nil {
		return false, fmt.Errorf("expire room: %w", tx.Error)
	}

	return tx.RowsAffected > 0, nil
}

func (r *Repository) CreatePlayer(ctx context.Context, roomID uuid.UUID, name string, lat, lng float64, orderIndex int) (*database.PlayerEntity, error) {
	var player database.PlayerEntity
	player.ID = uuid.New()
	player.RoomID = roomID
	player.Name = name
	// PostGIS は EWKT 形式を geography 型にキャストできる（経度, 緯度 の順）
	player.Location = fmt.Sprintf("SRID=4326;POINT(%f %f)", lng, lat)
	player.OrderIndex = orderIndex
	if err := r.db.WithContext(ctx).Create(&player).Error; err != nil {
		return nil, fmt.Errorf("create player: %w", err)
	}

	return &player, nil
}
