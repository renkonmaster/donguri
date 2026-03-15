package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gotest.tools/v3/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoomRepoForTest(t *testing.T) *Repository {
	t.Helper()

	// t.Name() をそのままファイル名に使うとスラッシュが入るため UUID で一意化
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{}) 
	assert.NilError(t, err)

	// geography(Point,4326) は SQLite 非対応のため TEXT で代替
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

	return New(db)
}

func TestFindMatchingRoom(t *testing.T) {
	t.Parallel()

	t.Run("not found returns nil", func(t *testing.T) {
		t.Parallel()
		repo := setupRoomRepoForTest(t)

		room, err := repo.FindMatchingRoom(context.Background(), 4)
		assert.NilError(t, err)
		assert.Assert(t, room == nil)
	})

	t.Run("returns room when not full", func(t *testing.T) {
		t.Parallel()
		repo := setupRoomRepoForTest(t)
		ctx := context.Background()

		room, err := repo.CreateRoom(ctx)
		assert.NilError(t, err)

		_, err = repo.CreatePlayer(ctx, room.ID, "u1", 33.0, 131.0, 0)
		assert.NilError(t, err)

		found, err := repo.FindMatchingRoom(ctx, 4)
		assert.NilError(t, err)
		assert.Assert(t, found != nil)
		assert.Equal(t, found.ID, room.ID)
	})

	t.Run("does not return full room", func(t *testing.T) {
		t.Parallel()
		repo := setupRoomRepoForTest(t)
		ctx := context.Background()

		room, err := repo.CreateRoom(ctx)
		assert.NilError(t, err)

		for i := range 4 {
			_, err := repo.CreatePlayer(ctx, room.ID, "u", 33.0, 131.0, i)
			assert.NilError(t, err)
		}

		found, err := repo.FindMatchingRoom(ctx, 4)
		assert.NilError(t, err)
		assert.Assert(t, found == nil)
	})
}

func TestCreatePlayer_SavesLocationAsEWKT(t *testing.T) {
	t.Parallel()

	repo := setupRoomRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	player, err := repo.CreatePlayer(ctx, room.ID, "user1", 33.2, 131.5, 0)
	assert.NilError(t, err)
	assert.Assert(t, player.ID != uuid.Nil)
	assert.Equal(t, player.Location, "SRID=4326;POINT(131.500000 33.200000)")
	assert.Equal(t, player.OrderIndex, 0)
}

func TestCountPlayersInRoom(t *testing.T) {
	t.Parallel()

	repo := setupRoomRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	_, err = repo.CreatePlayer(ctx, room.ID, "u1", 33.0, 131.0, 0)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "u2", 33.1, 131.1, 1)
	assert.NilError(t, err)

	count, err := repo.CountPlayersInRoom(ctx, room.ID)
	assert.NilError(t, err)
	assert.Equal(t, count, 2)
}

func TestMarkRoomPlayingIfFull(t *testing.T) {
	t.Parallel()

	t.Run("below capacity does not update", func(t *testing.T) {
		t.Parallel()
		repo := setupRoomRepoForTest(t)
		ctx := context.Background()

		room, err := repo.CreateRoom(ctx)
		assert.NilError(t, err)

		changed, err := repo.MarkRoomPlayingIfFull(ctx, room.ID, 3, 4, 10*time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, changed, false)

		var reloaded database.RoomEntity
		assert.NilError(t, repo.db.First(&reloaded, "id = ?", room.ID).Error)
		assert.Equal(t, reloaded.Status, database.RoomStatusMatching)
		assert.Assert(t, reloaded.StartAt == nil)
		assert.Assert(t, reloaded.ExpiresAt == nil)
	})

	t.Run("at capacity updates status and timestamps", func(t *testing.T) {
		t.Parallel()
		repo := setupRoomRepoForTest(t)
		ctx := context.Background()

		room, err := repo.CreateRoom(ctx)
		assert.NilError(t, err)

		changed, err := repo.MarkRoomPlayingIfFull(ctx, room.ID, 4, 4, 10*time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, changed, true)

		var reloaded database.RoomEntity
		assert.NilError(t, repo.db.First(&reloaded, "id = ?", room.ID).Error)
		assert.Equal(t, reloaded.Status, database.RoomStatusPlaying)
		assert.Assert(t, reloaded.StartAt != nil)
		assert.Assert(t, reloaded.ExpiresAt != nil)
		assert.Assert(t, reloaded.ExpiresAt.After(*reloaded.StartAt))
	})
}
