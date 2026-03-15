package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"
)

func setupRoomStateRepoForTest(t *testing.T) *Repository {
	t.Helper()

	repo := setupRoomRepoForTest(t)
	assert.NilError(t, repo.db.Exec(`
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

	return repo
}

func TestGetRoomState_RoomNotFound(t *testing.T) {
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
	ctx := context.Background()

	state, err := repo.GetRoomState(ctx, uuid.New(), uuid.New())
	assert.NilError(t, err)
	assert.Assert(t, state == nil)
}

func TestGetRoomState_MatchingRoom(t *testing.T) {
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
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
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
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
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
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
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)
	p2, err := repo.CreatePlayer(ctx, room.ID, "p2", 33.2, 131.2, 2)
	assert.NilError(t, err)

	// p0 → p1 に intent (p0 視点: my)
	// p2 → p0 に intent (p0 視点: received)
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap) VALUES (?, ?, ?, TRUE)`,
		room.ID, p0.ID, p1.ID,
	).Error)
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap) VALUES (?, ?, ?, TRUE)`,
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
	t.Parallel()

	repo := setupRoomStateRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	p0, err := repo.CreatePlayer(ctx, room.ID, "p0", 33.0, 131.0, 0)
	assert.NilError(t, err)
	p1, err := repo.CreatePlayer(ctx, room.ID, "p1", 33.1, 131.1, 1)
	assert.NilError(t, err)

	// needs_swap = FALSE のレコードは無視されるべき
	assert.NilError(t, repo.db.Exec(
		`INSERT INTO connections (room_id, sender_id, receiver_id, needs_swap) VALUES (?, ?, ?, FALSE)`,
		room.ID, p0.ID, p1.ID,
	).Error)

	state, err := repo.GetRoomState(ctx, room.ID, p0.ID)
	assert.NilError(t, err)
	assert.Equal(t, len(state.MyIntents), 0)
	assert.Equal(t, len(state.ReceivedIntents), 0)
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
