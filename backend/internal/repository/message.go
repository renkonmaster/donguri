package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
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
