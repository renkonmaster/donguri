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

type CreateMessageParams struct {
	RoomID     uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
}

func (r *Repository) CreateMessage(ctx context.Context, params CreateMessageParams) (*database.MessageEntity, error) {
	receiverID := params.ReceiverID
	message := &database.MessageEntity{
		ID:         uuid.New(),
		RoomID:     params.RoomID,
		SenderID:   params.SenderID,
		ReceiverID: &receiverID,
		Content:    params.Content,
		CreatedAt:  time.Now().UTC(),
	}

	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	return message, nil
}

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

type roomIntersectionCountRow struct {
	Count int `gorm:"column:count"`
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
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "room_id"},
				{Name: "sender_id"},
				{Name: "receiver_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"needs_swap": params.NeedsSwap,
				"updated_at": gorm.Expr("NOW()"),
			}),
		}).Create(intent).Error; err != nil {
			return fmt.Errorf("upsert swap intent: %w", err)
		}

		var room database.RoomEntity
		if err := tx.Select("status").Take(&room, "id = ?", params.RoomID).Error; err == nil {
			result.RoomStatus = room.Status
		}

		var reverseIntentCount int64
		if err := tx.Model(&database.ConnectionEntity{}).
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
		if err := tx.Model(&database.PlayerEntity{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
		if err := tx.Model(&database.PlayerEntity{}).
			Where("room_id = ?", params.RoomID).
			Select("COALESCE(MAX(order_index), 0)").
			Scan(&maxOrderIndex).Error; err != nil {
			return fmt.Errorf("select max order index: %w", err)
		}

		tempOrderIndex := maxOrderIndex + 1
		if err := tx.Model(&database.PlayerEntity{}).
			Where("room_id = ? AND id = ?", params.RoomID, params.SenderID).
			Updates(map[string]any{"order_index": tempOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move sender to temp order index: %w", err)
		}

		if err := tx.Model(&database.PlayerEntity{}).
			Where("room_id = ? AND id = ?", params.RoomID, params.ReceiverID).
			Updates(map[string]any{"order_index": senderOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move receiver to sender order index: %w", err)
		}

		if err := tx.Model(&database.PlayerEntity{}).
			Where("room_id = ? AND id = ?", params.RoomID, params.SenderID).
			Updates(map[string]any{"order_index": receiverOrderIndex, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return fmt.Errorf("move sender to receiver order index: %w", err)
		}

		if err := tx.Where(
			"room_id = ? AND (sender_id IN ? OR receiver_id IN ?)",
			params.RoomID,
			[]uuid.UUID{params.SenderID, params.ReceiverID},
			[]uuid.UUID{params.SenderID, params.ReceiverID},
		).Delete(&database.ConnectionEntity{}).Error; err != nil {
			return fmt.Errorf("delete affected connections: %w", err)
		}

		var intersectionCountRow roomIntersectionCountRow
		if err := tx.Model(&database.RoomEntity{}).
			Select("get_room_intersection_count(?) AS count", params.RoomID).
			Limit(1).
			Scan(&intersectionCountRow).Error; err != nil {
			return fmt.Errorf("get room intersection count: %w", err)
		}

		if intersectionCountRow.Count != 0 {
			return nil
		}

		if err := tx.Model(&database.RoomEntity{}).
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
