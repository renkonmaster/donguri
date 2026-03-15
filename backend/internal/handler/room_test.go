package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"
	"gotest.tools/v3/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGetRoomHandler(t *testing.T) *Handler {
	t.Helper()

	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{}) //nolint:exhaustruct
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

	assert.NilError(t, db.Exec(`
		CREATE TABLE connections (
			room_id     TEXT NOT NULL REFERENCES rooms(id),
			sender_id   TEXT NOT NULL REFERENCES players(id),
			receiver_id TEXT NOT NULL REFERENCES players(id),
			needs_swap  BOOLEAN NOT NULL DEFAULT FALSE,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (room_id, sender_id, receiver_id),
			CHECK (sender_id <> receiver_id)
		)
	`).Error)

	return New(repository.New(db), stream.NewHub())
}

func TestGetRoom_NotFound(t *testing.T) {
	t.Parallel()

	h := setupGetRoomHandler(t)

	_, err := h.GetRoom(context.Background(), api.GetRoomParams{
		RoomID:    uuid.New(),
		XPlayerID: uuid.New(),
	})

	apiErr, ok := err.(*api.ErrorStatusCode)
	assert.Assert(t, ok, "expected *api.ErrorStatusCode")
	assert.Equal(t, apiErr.StatusCode, http.StatusNotFound)
}

func TestGetRoom_ReturnsState(t *testing.T) {
	t.Parallel()

	h := setupGetRoomHandler(t)
	ctx := context.Background()

	joinRes, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p0", Lat: 33.0, Lng: 131.0})
	assert.NilError(t, err)

	res, err := h.GetRoom(ctx, api.GetRoomParams{
		RoomID:    joinRes.RoomID,
		XPlayerID: joinRes.PlayerID,
	})
	assert.NilError(t, err)
	assert.Assert(t, res != nil)

	assert.Equal(t, res.Status, api.RoomStatusMatching)
	assert.Assert(t, res.TimeLeft.IsNull())
	assert.Equal(t, len(res.Players), 1)
	assert.Equal(t, res.Players[0].ID, joinRes.PlayerID)
	assert.Equal(t, res.Players[0].Name, "p0")
	assert.Equal(t, res.Players[0].Location.Lat, 33.0)
	assert.Equal(t, res.Players[0].Location.Lng, 131.0)
}

func TestGetRoom_PlayersOrderedByIndex(t *testing.T) {
	t.Parallel()

	h := setupGetRoomHandler(t)
	ctx := context.Background()

	r0, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p0", Lat: 33.0, Lng: 131.0})
	assert.NilError(t, err)
	r1, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p1", Lat: 33.1, Lng: 131.1})
	assert.NilError(t, err)
	assert.Equal(t, r0.RoomID, r1.RoomID)

	res, err := h.GetRoom(ctx, api.GetRoomParams{
		RoomID:    r0.RoomID,
		XPlayerID: r0.PlayerID,
	})
	assert.NilError(t, err)
	assert.Equal(t, len(res.Players), 2)
	assert.Equal(t, res.Players[0].OrderIndex, 0)
	assert.Equal(t, res.Players[1].OrderIndex, 1)
}

func TestGetRoom_IntentsReflected(t *testing.T) {
	t.Parallel()

	h := setupGetRoomHandler(t)
	ctx := context.Background()

	r0, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p0", Lat: 33.0, Lng: 131.0})
	assert.NilError(t, err)
	r1, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p1", Lat: 33.1, Lng: 131.1})
	assert.NilError(t, err)
	assert.Equal(t, r0.RoomID, r1.RoomID)

	// p0 → p1 に intent を直接挿入
	assert.NilError(t, h.repo.(*repository.Repository).ExecRaw(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap) VALUES (?, ?, ?, TRUE)`,
		r0.RoomID, r0.PlayerID, r1.PlayerID,
	))

	// p0 視点: my_intents に p1 が入る
	res0, err := h.GetRoom(ctx, api.GetRoomParams{RoomID: r0.RoomID, XPlayerID: r0.PlayerID})
	assert.NilError(t, err)
	assert.Equal(t, len(res0.MyIntents), 1)
	assert.Equal(t, res0.MyIntents[0], r1.PlayerID)
	assert.Equal(t, len(res0.ReceivedIntents), 0)

	// p1 視点: received_intents に p0 が入る
	res1, err := h.GetRoom(ctx, api.GetRoomParams{RoomID: r0.RoomID, XPlayerID: r1.PlayerID})
	assert.NilError(t, err)
	assert.Equal(t, len(res1.ReceivedIntents), 1)
	assert.Equal(t, res1.ReceivedIntents[0], r0.PlayerID)
	assert.Equal(t, len(res1.MyIntents), 0)
}
