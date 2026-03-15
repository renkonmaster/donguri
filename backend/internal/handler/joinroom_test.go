package handler

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gotest.tools/v3/assert"
)

func setupJoinRoomHandler(t *testing.T) *Handler {
	t.Helper()

	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{}) 
	assert.NilError(t, err)

	assert.NilError(t, db.Exec(`
		CREATE TABLE rooms (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			start_at DATETIME,
			expires_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)

	assert.NilError(t, db.Exec(`
		CREATE TABLE players (
			id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL REFERENCES rooms(id),
			name TEXT NOT NULL,
			location TEXT NOT NULL,
			order_index INTEGER NOT NULL CHECK (order_index >= 0),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (room_id, order_index)
		)
	`).Error)

	repo := repository.New(db)
	hub := stream.NewHub()

	return New(repo, hub)
}


func TestJoinRoom_ReturnsIDs(t *testing.T) {
    t.Parallel()

    h := setupJoinRoomHandler(t)

    res, err := h.JoinRoom(context.Background(), &api.JoinRoomRequest{
        Name: "user1",
        Lat:  33.2,
        Lng:  131.5,
    })
    assert.NilError(t, err)
    assert.Assert(t, res != nil)
    assert.Assert(t, res.RoomID != uuid.Nil)
    assert.Assert(t, res.PlayerID != uuid.Nil)
}

func TestJoinRoom_PublishesRoomUpdatedEvent(t *testing.T) {
    t.Parallel()

    h := setupJoinRoomHandler(t)

    first, err := h.JoinRoom(context.Background(), &api.JoinRoomRequest{
        Name: "u1",
        Lat:  33.0,
        Lng:  131.0,
    })
    assert.NilError(t, err)

    sub := stream.NewSubscriber("test-sub", 1)
    h.hub.Subscribe(first.RoomID.String(), sub)
    defer h.hub.Unsubscribe(first.RoomID.String(), sub.ID)

    _, err = h.JoinRoom(context.Background(), &api.JoinRoomRequest{
        Name: "u2",
        Lat:  33.1,
        Lng:  131.1,
    })
    assert.NilError(t, err)

    select {
    case payload := <-sub.Ch:
        s := string(payload)
        assert.Assert(t, strings.Contains(s, "event: room_updated"))
        assert.Assert(t, strings.Contains(s, "\"joined_count\":2"))
        assert.Assert(t, strings.Contains(s, "\"event\":\"joined\""))
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting room_updated event")
    }
}