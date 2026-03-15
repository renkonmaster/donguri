package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SetSwapIntentParams struct {
	RoomID     uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	NeedsSwap  bool
}

type SetSwapIntentResult struct {
	Matched    bool
	RoomStatus string
}

func (r *Repository) SetSwapIntent(ctx context.Context, params SetSwapIntentParams) (*SetSwapIntentResult, error) {
	now := time.Now().UTC()
	intent := &database.ConnectionEntity{
		RoomID:     params.RoomID,
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
		NeedsSwap:  params.NeedsSwap,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result := &SetSwapIntentResult{Matched: false, RoomStatus: database.RoomStatusPlaying}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var room database.RoomEntity
		if err := tx.Select("status").Take(&room, "id = ?", params.RoomID).Error; err != nil {
			return fmt.Errorf("select room: %w", err)
		}
		result.RoomStatus = room.Status
		if room.Status != database.RoomStatusPlaying {
			return ErrRoomNotPlaying
		}

		var conflict clause.OnConflict
		var c1, c2, c3 clause.Column
		c1.Name = "room_id"
		c2.Name = "sender_id"
		c3.Name = "receiver_id"
		conflict.Columns = []clause.Column{c1, c2, c3}
		conflict.DoUpdates = clause.Assignments(map[string]any{
			"needs_swap": params.NeedsSwap,
			"updated_at": gorm.Expr("NOW()"),
		})
		if err := tx.Clauses(conflict).Create(intent).Error; err != nil {
			return fmt.Errorf("upsert swap intent: %w", err)
		}

		var reverseIntentCount int64
		if err := tx.Model(new(database.ConnectionEntity)).
			Where(
				"room_id = ? AND sender_id = ? AND receiver_id = ? AND needs_swap = TRUE",
				params.RoomID,
				params.ReceiverID,
				params.SenderID,
			).
			Count(&reverseIntentCount).Error; err != nil {
			return fmt.Errorf("select reverse swap intent: %w", err)
		}

		if reverseIntentCount == 0 {
			return nil
		}

		result.Matched = true

		playerIDs := []uuid.UUID{params.SenderID, params.ReceiverID}
		var swapPlayers []database.PlayerEntity
		var locking clause.Locking
		locking.Strength = "UPDATE"
		if err := tx.Model(new(database.PlayerEntity)).
			Clauses(locking).
			Where("room_id = ? AND id IN ?", params.RoomID, playerIDs).
			Find(&swapPlayers).Error; err != nil {
			return fmt.Errorf("select players for swap: %w", err)
		}
		if len(swapPlayers) != 2 {
			return fmt.Errorf("swap targets not found in room")
		}

		var senderOrderIndex, receiverOrderIndex int
		for _, player := range swapPlayers {
			switch player.ID {
			case params.SenderID:
				senderOrderIndex = player.OrderIndex
			case params.ReceiverID:
				receiverOrderIndex = player.OrderIndex
			}
		}

		var maxOrderIndex int
		if err := tx.Model(new(database.PlayerEntity)).
			Where("room_id = ?", params.RoomID).
			Select("COALESCE(MAX(order_index), 0)").
			Scan(&maxOrderIndex).Error; err != nil {
			return fmt.Errorf("select max order index: %w", err)
		}

		tempOrderIndex := maxOrderIndex + 1
		if err := tx.Model(new(database.PlayerEntity)).
			Where("room_id = ? AND id = ?", params.RoomID, params.SenderID).
			Updates(map[string]any{"order_index": tempOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move sender to temp order index: %w", err)
		}

		if err := tx.Model(new(database.PlayerEntity)).
			Where("room_id = ? AND id = ?", params.RoomID, params.ReceiverID).
			Updates(map[string]any{"order_index": senderOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move receiver to sender order index: %w", err)
		}

		if err := tx.Model(new(database.PlayerEntity)).
			Where("room_id = ? AND id = ?", params.RoomID, params.SenderID).
			Updates(map[string]any{"order_index": receiverOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move sender to receiver order index: %w", err)
		}

		if err := tx.Where(
			"room_id = ? AND (sender_id IN ? OR receiver_id IN ?)",
			params.RoomID,
			[]uuid.UUID{params.SenderID, params.ReceiverID},
			[]uuid.UUID{params.SenderID, params.ReceiverID},
		).Delete(new(database.ConnectionEntity)).Error; err != nil {
			return fmt.Errorf("delete affected connections: %w", err)
		}

		intersectionCount, err := r.GetRoomIntersectionCount(ctx, params.RoomID, tx)
		if err != nil {
			return err
		}
		if intersectionCount != 0 {
			return nil
		}
		if err := tx.Model(new(database.RoomEntity)).
			Where("id = ?", params.RoomID).
			Updates(map[string]any{"status": database.RoomStatusFinished, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("update room status to finished: %w", err)
		}
		result.RoomStatus = database.RoomStatusFinished

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
