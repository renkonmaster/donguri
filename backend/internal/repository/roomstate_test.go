package repository

// GetRoomState は PostGIS 専用の GetRoomIntersectionCount を呼ぶため、
// POSTGRES_TEST_DSN 環境変数を設定した実 PostgreSQL+PostGIS が必要。
//
// 実行例:
//   POSTGRES_TEST_DSN="postgres://user:password@localhost:5432/database?sslmode=disable" \
//   go test ./internal/repository/ -run TestGetRoomState -v

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gotest.tools/v3/assert"
)

func TestGetRoomState_RoomNotFound(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	state, err := repo.GetRoomState(ctx, uuid.New(), uuid.New())
	assert.NilError(t, err)
	assert.Assert(t, state == nil)
}

func TestGetRoomState_MatchingRoom(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)

	state, err := repo.GetRoomState(ctx, room.ID, p0.ID)
	assert.NilError(t, err)
	assert.Assert(t, state != nil)

	assert.Equal(t, state.Status, "matching")
	assert.Assert(t, state.TimeLeftSeconds == nil)
	assert.Equal(t, len(state.Players), 2)
	assert.Equal(t, state.Players[0].ID, p0.ID)
	assert.Equal(t, state.Players[1].ID, p1.ID)
	assert.Equal(t, len(state.MyIntents), 0)
	assert.Equal(t, len(state.ReceivedIntents), 0)
}

func TestGetRoomState_PlayersOrderedByIndex(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	// 逆順で登録してもorder_index昇順で返ることを確認
	p2, err := repo.CreatePlayer(ctx, room.ID, "p2", 33.2, 131.2, 2)
	assert.NilError(t, err)
	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)

	state, err := repo.GetRoomState(ctx, room.ID, p0.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(state.Players), 3)
	assert.Equal(t, state.Players[0].ID, p0.ID)
	assert.Equal(t, state.Players[1].ID, p1.ID)
	assert.Equal(t, state.Players[2].ID, p2.ID)
}

func TestGetRoomState_PlayerLocationRoundTrip(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.2, 131.5, 0)
	assert.NilError(t, err)

	state, err := repo.GetRoomState(ctx, room.ID, p.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(state.Players), 1)
	assert.Equal(t, state.Players[0].Lat, 33.2)
	assert.Equal(t, state.Players[0].Lng, 131.5)
}

func TestGetRoomState_MyAndReceivedIntents(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)
	p2, err := repo.CreatePlayer(ctx, room.ID, "p2", 33.2, 131.2, 2)
	assert.NilError(t, err)

	// p0 → p1 (p0 視点: my)、p2 → p0 (p0 視点: received)
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap, created_at, updated_at) VALUES (?, ?, ?, TRUE, NOW(), NOW())`,
		room.ID, p0.ID, p1.ID,
	).Error)
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap, created_at, updated_at) VALUES (?, ?, ?, TRUE, NOW(), NOW())`,
		room.ID, p2.ID, p0.ID,
	).Error)

	state, err := repo.GetRoomState(ctx, room.ID, p0.ID)
	assert.NilError(t, err)

	assert.Equal(t, len(state.MyIntents), 1)
	assert.Equal(t, state.MyIntents[0], p1.ID)
	assert.Equal(t, len(state.ReceivedIntents), 1)
	assert.Equal(t, state.ReceivedIntents[0], p2.ID)
}

func TestGetRoomState_NeedsSwapFalseNotIncluded(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)

	// needs_swap = FALSE のレコードは intent として無視される
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap, created_at, updated_at) VALUES (?, ?, ?, FALSE, NOW(), NOW())`,
		room.ID, p0.ID, p1.ID,
	).Error)

	state, err := repo.GetRoomState(ctx, room.ID, p0.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(state.MyIntents), 0)
	assert.Equal(t, len(state.ReceivedIntents), 0)
}

func TestGetRoomState_PlayingRoomHasTimeLeft(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)

	// 直接 playing に更新して expires_at をセット
	assert.NilError(t, repo.db.Exec(
		`UPDATE rooms SET status = ?, start_at = NOW(), expires_at = NOW() + INTERVAL '5 minutes', updated_at = NOW() WHERE id = ?`,
		database.RoomStatusPlaying, room.ID,
	).Error)

	state, err := repo.GetRoomState(ctx, room.ID, p.ID)
	assert.NilError(t, err)
	assert.Equal(t, state.Status, database.RoomStatusPlaying)
	assert.Assert(t, state.TimeLeftSeconds != nil)
	assert.Assert(t, *state.TimeLeftSeconds > 0)
}

func TestParseEWKTPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantLat float64
		wantLng float64
	}{
		{"SRID=4326;POINT(131.5 33.2)", 33.2, 131.5},
		{"POINT(0 0)", 0, 0},
		{"SRID=4326;POINT(-122.4194 37.7749)", 37.7749, -122.4194},
	}

	for _, tt := range tests {
		got, err := parseEWKTPoint(tt.input)
		assert.NilError(t, err)
		assert.Equal(t, got.Lat, tt.wantLat)
		assert.Equal(t, got.Lng, tt.wantLng)
	}
}

func TestParseEWKTPoint_InvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := parseEWKTPoint("invalid_format")
	assert.Assert(t, err != nil)
}
