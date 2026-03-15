package repository

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"gorm.io/gorm"
)

// ShufflePlayerOrderIndices はルーム内のプレイヤーの order_index をシャッフルする。
// 0 回か 1 回のスワップでクリアできる配置の場合は再シャッフルする。
// 有効な配置が見つからない場合は警告を出してそのまま返す。
func (r *Repository) ShufflePlayerOrderIndices(ctx context.Context, roomID uuid.UUID) error {
	type playerRow struct {
		ID  uuid.UUID `gorm:"column:id"`
		Lat float64   `gorm:"column:lat"`
		Lng float64   `gorm:"column:lng"`
	}
	var rows []playerRow
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id, ST_Y(location::geometry) AS lat, ST_X(location::geometry) AS lng
		 FROM players WHERE room_id = ? ORDER BY order_index`,
		roomID,
	).Scan(&rows).Error; err != nil {
		return fmt.Errorf("fetch players for shuffle: %w", err)
	}

	n := len(rows)
	// 4 点未満では線分の交差が原理的に起きない
	if n < 4 {
		return nil
	}

	locs := make([]Point, n)
	ids := make([]uuid.UUID, n)
	for i, row := range rows {
		locs[i] = Point{Lat: row.Lat, Lng: row.Lng}
		ids[i] = row.ID
	}

	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	valid := false
	const maxAttempts = 100
	for range maxAttempts {
		rand.Shuffle(n, func(i, j int) {
			perm[i], perm[j] = perm[j], perm[i]
		})

		pts := make([]Point, n)
		for i, pi := range perm {
			pts[i] = locs[pi]
		}

		if countIntersectionsInMem(pts) == 0 || solvableInOneSwap(pts) {
			continue
		}

		valid = true
		break
	}

	if !valid {
		slog.Warn("ShufflePlayerOrderIndices: valid shuffle not found, applying last attempt", "room_id", roomID)
	}
	return applyOrderIndexPermutation(ctx, r.db, roomID, ids, perm)
}

// applyOrderIndexPermutation は ids[perm[i]] に order_index=i を割り当てる。
// unique 制約の競合を避けるため、一度 n..2n-1 の仮値にシフトしてから最終値を設定する。
func applyOrderIndexPermutation(ctx context.Context, db *gorm.DB, roomID uuid.UUID, ids []uuid.UUID, perm []int) error {
	n := len(ids)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(new(database.PlayerEntity)).
			Where("room_id = ?", roomID).
			Updates(map[string]any{
				"order_index": gorm.Expr("order_index + ?", n),
				"updated_at":  gorm.Expr("NOW()"),
			}).Error; err != nil {
			return fmt.Errorf("shift order indices to temp range: %w", err)
		}

		for i, pi := range perm {
			if err := tx.Model(new(database.PlayerEntity)).
				Where("room_id = ? AND id = ?", roomID, ids[pi]).
				Updates(map[string]any{
					"order_index": i,
					"updated_at":  gorm.Expr("NOW()"),
				}).Error; err != nil {
				return fmt.Errorf("set order_index %d: %w", i, err)
			}
		}
		return nil
	})
}

// countIntersectionsInMem は points を order_index 順に結んだ折れ線の交差数を返す。
func countIntersectionsInMem(pts []Point) int {
	n := len(pts)
	count := 0
	for i := 0; i < n-2; i++ {
		for j := i + 2; j < n-1; j++ {
			if segmentsIntersect(pts[i], pts[i+1], pts[j], pts[j+1]) {
				count++
			}
		}
	}
	return count
}

// solvableInOneSwap は任意の 1 スワップで交差数が 0 になるかを判定する。
func solvableInOneSwap(pts []Point) bool {
	n := len(pts)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			pts[i], pts[j] = pts[j], pts[i]
			solved := countIntersectionsInMem(pts) == 0
			pts[i], pts[j] = pts[j], pts[i]
			if solved {
				return true
			}
		}
	}
	return false
}

// segmentsIntersect は線分 p1-p2 と p3-p4 が（端点接触を除いて）交差するか判定する。
func segmentsIntersect(p1, p2, p3, p4 Point) bool {
	d1 := crossProduct(p3, p4, p1)
	d2 := crossProduct(p3, p4, p2)
	d3 := crossProduct(p1, p2, p3)
	d4 := crossProduct(p1, p2, p4)
	return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))
}

// crossProduct は o を原点としたベクトル (o→a) × (o→b) の z 成分を返す。
func crossProduct(o, a, b Point) float64 {
	return (a.Lng-o.Lng)*(b.Lat-o.Lat) - (a.Lat-o.Lat)*(b.Lng-o.Lng)
}
