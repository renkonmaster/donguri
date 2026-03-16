package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gotest.tools/v3/assert"
)

func setupMessageRepoForTest(t *testing.T) *Repository {
	t.Helper()

	repo := setupRoomRepoForTest(t)
	assert.NilError(t, repo.db.Exec(`
		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL REFERENCES rooms(id),
			sender_id TEXT NOT NULL REFERENCES players(id),
			receiver_id TEXT REFERENCES players(id),
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`).Error)

	return repo
}

func TestCreateMessage_AdjacentValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := setupMessageRepoForTest(t)

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)
	assert.NilError(t, repo.db.WithContext(ctx).Model(new(database.RoomEntity)).
		Where("id = ?", room.ID).
		Update("status", database.RoomStatusPlaying).Error)

	// 4 人のリングを構成: 0-1-2-3-0。0 と 1 は隣接、0 と 2 は非隣接
	sender, err := repo.CreatePlayer(ctx, room.ID, "sender", 33.0, 131.0, 0)
	assert.NilError(t, err)
	adjacent, err := repo.CreatePlayer(ctx, room.ID, "adjacent", 33.1, 131.1, 1)
	assert.NilError(t, err)
	notAdjacent, err := repo.CreatePlayer(ctx, room.ID, "not-adjacent", 33.2, 131.2, 2)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p3", 33.3, 131.3, 3)
	assert.NilError(t, err)

	t.Run("success when order_index diff is 1", func(t *testing.T) {
		message, createErr := repo.CreateMessage(ctx, CreateMessageParams{
			RoomID:     room.ID,
			SenderID:   sender.ID,
			ReceiverID: adjacent.ID,
			Content:    "hello",
		})
		assert.NilError(t, createErr)
		assert.Assert(t, message != nil)
		assert.Assert(t, message.ReceiverID != nil)
		assert.Equal(t, *message.ReceiverID, adjacent.ID)
	})

	t.Run("success when order_index diff is n-1 (wrap-around)", func(t *testing.T) {
		// sender (0) と p3 (3) は diff=3=n-1 で隣接
		var p3Entity database.PlayerEntity
		assert.NilError(t, repo.db.WithContext(ctx).Where("room_id = ? AND name = ?", room.ID, "p3").First(&p3Entity).Error)
		message, createErr := repo.CreateMessage(ctx, CreateMessageParams{
			RoomID:     room.ID,
			SenderID:   sender.ID,
			ReceiverID: p3Entity.ID,
			Content:    "wrap-around",
		})
		assert.NilError(t, createErr)
		assert.Assert(t, message != nil)
	})

	t.Run("fails when order_index diff is not adjacent", func(t *testing.T) {
		// sender (0) と notAdjacent (2) は diff=2、n=4 なので隣接しない
		_, createErr := repo.CreateMessage(ctx, CreateMessageParams{
			RoomID:     room.ID,
			SenderID:   sender.ID,
			ReceiverID: notAdjacent.ID,
			Content:    "cannot send",
		})
		assert.Assert(t, errors.Is(createErr, ErrPlayersNotAdjacent))
	})

	t.Run("fails when receiver is not in same room", func(t *testing.T) {
		otherRoom, createErr := repo.CreateRoom(ctx)
		assert.NilError(t, createErr)
		otherPlayer, createErr := repo.CreatePlayer(ctx, otherRoom.ID, "other", 34.0, 132.0, 0)
		assert.NilError(t, createErr)

		_, createErr = repo.CreateMessage(ctx, CreateMessageParams{
			RoomID:     room.ID,
			SenderID:   sender.ID,
			ReceiverID: otherPlayer.ID,
			Content:    "cannot send",
		})
		assert.Assert(t, errors.Is(createErr, ErrPlayerNotFoundInRoom))
	})

	t.Run("fails when sender and receiver are same", func(t *testing.T) {
		_, createErr := repo.CreateMessage(ctx, CreateMessageParams{
			RoomID:     room.ID,
			SenderID:   sender.ID,
			ReceiverID: sender.ID,
			Content:    "cannot send",
		})
		assert.Assert(t, errors.Is(createErr, ErrSenderReceiverSame))
	})
}

func TestGetMessagesRelatedToPlayer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := setupMessageRepoForTest(t)

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)
	p2, err := repo.CreatePlayer(ctx, room.ID, "p2", 33.2, 131.2, 2)
	assert.NilError(t, err)

	now := time.Now().UTC()
	messages := []database.MessageEntity{
		{
			ID:         uuid.New(),
			RoomID:     room.ID,
			SenderID:   p0.ID,
			ReceiverID: &p1.ID,
			Content:    "m1",
			CreatedAt:  now.Add(-3 * time.Second),
		},
		{
			ID:         uuid.New(),
			RoomID:     room.ID,
			SenderID:   p1.ID,
			ReceiverID: &p0.ID,
			Content:    "m2",
			CreatedAt:  now.Add(-2 * time.Second),
		},
		{
			ID:         uuid.New(),
			RoomID:     room.ID,
			SenderID:   p2.ID,
			ReceiverID: &p1.ID,
			Content:    "m3",
			CreatedAt:  now.Add(-1 * time.Second),
		},
	}
	for _, message := range messages {
		assert.NilError(t, repo.db.WithContext(ctx).Create(&message).Error)
	}

	related, err := repo.GetMessagesRelatedToPlayer(ctx, room.ID, p1.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(related), 3)

	otherRelated, err := repo.GetMessagesRelatedToPlayer(ctx, room.ID, p2.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(otherRelated), 1)
	assert.Equal(t, otherRelated[0].Content, "m3")
}
