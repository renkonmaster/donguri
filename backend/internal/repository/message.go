package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
)

var (
	ErrReceiverIDRequired   = errors.New("receiver_id is required")
	ErrSenderReceiverSame   = errors.New("sender and receiver must be different")
	ErrPlayerNotFoundInRoom = errors.New("player is not found in room")
	ErrPlayersNotAdjacent   = errors.New("sender and receiver are not adjacent")
)

type CreateMessageParams struct {
	RoomID     uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
}

func (r *Repository) CreateMessage(ctx context.Context, params CreateMessageParams) (*database.MessageEntity, error) {
	if params.ReceiverID == uuid.Nil {
		return nil, ErrReceiverIDRequired
	}

	if params.SenderID == params.ReceiverID {
		return nil, ErrSenderReceiverSame
	}

	var players []database.PlayerEntity
	if err := r.db.WithContext(ctx).
		Model(&database.PlayerEntity{}).
		Where("room_id = ? AND id IN ?", params.RoomID, []uuid.UUID{params.SenderID, params.ReceiverID}).
		Select("id", "order_index").
		Find(&players).Error; err != nil {
		return nil, fmt.Errorf("select sender/receiver players: %w", err)
	}
	if len(players) != 2 {
		return nil, ErrPlayerNotFoundInRoom
	}

	orderByPlayerID := make(map[uuid.UUID]int, 2)
	for _, player := range players {
		orderByPlayerID[player.ID] = player.OrderIndex
	}

	diff := orderByPlayerID[params.SenderID] - orderByPlayerID[params.ReceiverID]
	if diff < 0 {
		diff = -diff
	}
	if diff != 1 {
		return nil, ErrPlayersNotAdjacent
	}

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

func (r *Repository) GetMessagesRelatedToPlayer(ctx context.Context, roomID, playerID uuid.UUID) ([]database.MessageEntity, error) {
	var messages []database.MessageEntity

	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("(sender_id = ? OR receiver_id = ?)", playerID, playerID).
		Order("created_at ASC").
		Order("id ASC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("select related messages: %w", err)
	}

	return messages, nil
}
