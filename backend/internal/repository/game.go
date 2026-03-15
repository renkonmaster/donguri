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

func (r *Repository) SetSwapIntent(ctx context.Context, params SetSwapIntentParams) error {
	now := time.Now().UTC()
	intent := &database.ConnectionEntity{
		RoomID:     params.RoomID,
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
		NeedsSwap:  params.NeedsSwap,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "room_id"},
			{Name: "sender_id"},
			{Name: "receiver_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"needs_swap": params.NeedsSwap,
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(intent).Error
	if err != nil {
		return fmt.Errorf("upsert swap intent: %w", err)
	}

	return nil
}
