package repository

// ST_MakeLine / ST_Intersects / ST_Touches は PostGIS 専用関数のため、
// このテストは実 PostgreSQL+PostGIS が必要。
// POSTGRES_TEST_DSN 環境変数を設定しない場合は自動的にスキップする。
//
// 実行例:
//   POSTGRES_TEST_DSN="postgres://user:password@localhost:5432/database?sslmode=disable" \
//   go test ./internal/repository/ -run TestIntersection -v

import (
	"context"
	"math"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gotest.tools/v3/assert"
)

func setupPostGISRepoForTest(t *testing.T) *Repository {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostGIS test")
	}

	var config gorm.Config
	db, err := gorm.Open(postgres.Open(dsn), &config) 
	assert.NilError(t, err)

	repo := New(db)

	t.Cleanup(func() {
		db.Exec("DELETE FROM players")
		db.Exec("DELETE FROM rooms")
	})

	return repo
}

func TestGetIntersectingEdgePairs_Empty(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	pairs, err := repo.GetIntersectingEdgePairs(ctx, room.ID, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(pairs), 0)
}

func TestGetIntersectingEdgePairs_NonCrossing(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	// 正方形の角を順に結ぶ（C字型:交差なし）
	// 辺: 0→1 (lat=0,lng=0 → lat=0,lng=1)
	//      1→2 (lat=0,lng=1 → lat=1,lng=1)
	//      2→3 (lat=1,lng=1 → lat=1,lng=0)
	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	_, err = repo.CreatePlayer(ctx, room.ID, "p0", 0.0, 0.0, 0)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p1", 0.0, 1.0, 1)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p2", 1.0, 1.0, 2)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p3", 1.0, 0.0, 3)
	assert.NilError(t, err)

	pairs, err := repo.GetIntersectingEdgePairs(ctx, room.ID, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(pairs), 0)
}

func TestGetIntersectingEdgePairs_Crossing(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	// 交差する配置:
	//   辺0→1: POINT(0 0) → POINT(1 1)  対角線↗
	//   辺1→2: POINT(1 1) → POINT(1 0)  縦線↓
	//   辺2→3: POINT(1 0) → POINT(0 1)  対角線↖
	// → 辺0→1 と 辺2→3 が (0.5, 0.5) で交差する
	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	_, err = repo.CreatePlayer(ctx, room.ID, "p0", 0.0, 0.0, 0)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p1", 1.0, 1.0, 1)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p2", 0.0, 1.0, 2)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p3", 1.0, 0.0, 3)
	assert.NilError(t, err)

	pairs, err := repo.GetIntersectingEdgePairs(ctx, room.ID, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(pairs), 1)
	assert.Equal(t, pairs[0].First, Edge{StartOrderIndex: 0, EndOrderIndex: 1})
	assert.Equal(t, pairs[0].Second, Edge{StartOrderIndex: 2, EndOrderIndex: 3})
	assert.Assert(t, math.Abs(pairs[0].Location.Lat-0.5) < 1e-9)
	assert.Assert(t, math.Abs(pairs[0].Location.Lng-0.5) < 1e-9)
}

func TestGetRoomIntersectionCount_Crossing(t *testing.T) {
	repo := setupPostGISRepoForTest(t)
	ctx := context.Background()

	room, err := repo.CreateRoom(ctx)
	assert.NilError(t, err)

	_, err = repo.CreatePlayer(ctx, room.ID, "p0", 0.0, 0.0, 0)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p1", 1.0, 1.0, 1)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p2", 0.0, 1.0, 2)
	assert.NilError(t, err)
	_, err = repo.CreatePlayer(ctx, room.ID, "p3", 1.0, 0.0, 3)
	assert.NilError(t, err)

	count, err := repo.GetRoomIntersectionCount(ctx, room.ID, nil)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}
