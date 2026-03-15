package handler

// GetRoom ハンドラーのテストは PostGIS 専用の GetRoomIntersectionCount を呼ぶため、
// POSTGRES_TEST_DSN 環境変数を設定した実 PostgreSQL+PostGIS が必要。
//
// 実行例:
//   POSTGRES_TEST_DSN="postgres://user:password@localhost:5432/database?sslmode=disable" \
//   go test ./internal/handler/ -run TestGetRoom -v

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/repository"
	"github.com/renkonmaster/donguri/internal/service/stream"
	"gotest.tools/v3/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupGetRoomHandler(t *testing.T) *Handler {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping GetRoom handler test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) //nolint:exhaustruct
	assert.NilError(t, err)

	t.Cleanup(func() {
		db.Exec("DELETE FROM connections")
		db.Exec("DELETE FROM players")
		db.Exec("DELETE FROM rooms")
	})

	return New(repository.New(db), stream.NewHub())
}

func TestGetRoom_NotFound(t *testing.T) {
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
	h := setupGetRoomHandler(t)
	ctx := context.Background()

	r0, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p0", Lat: 33.0, Lng: 131.0})
	assert.NilError(t, err)
	r1, err := h.JoinRoom(ctx, &api.JoinRoomRequest{Name: "p1", Lat: 33.1, Lng: 131.1})
	assert.NilError(t, err)
	assert.Equal(t, r0.RoomID, r1.RoomID)

	// p0 → p1 に intent をセット
	_, err = h.repo.SetSwapIntent(ctx, repository.SetSwapIntentParams{
		RoomID:     r0.RoomID,
		SenderID:   r0.PlayerID,
		ReceiverID: r1.PlayerID,
		NeedsSwap:  true,
	})
	assert.NilError(t, err)

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
