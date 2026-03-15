package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Edge は order_index で表現した辺。
type Edge struct {
	StartOrderIndex int
	EndOrderIndex   int
}

// IntersectingEdgePair は交差している2辺の組。
type IntersectingEdgePair struct {
	First  Edge
	Second Edge
}

// GetIntersectingEdgePairs は room 内の交差している辺ペアを返す。
// ST_Touches=true（端点で接しているだけ）のケースは除外する。
func (r *Repository) GetIntersectingEdgePairs(ctx context.Context, roomID uuid.UUID, tx *gorm.DB) ([]IntersectingEdgePair, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	type row struct {
		FirstStart  int `gorm:"column:first_start"`
		FirstEnd    int `gorm:"column:first_end"`
		SecondStart int `gorm:"column:second_start"`
		SecondEnd   int `gorm:"column:second_end"`
	}
	var rows []row

	err := db.WithContext(ctx).Raw(`
		WITH lines AS (
			SELECT
				order_index AS start_order,
				LEAD(order_index) OVER (ORDER BY order_index) AS end_order,
				ST_MakeLine(
					ST_GeomFromEWKT(location),
					LEAD(ST_GeomFromEWKT(location)) OVER (ORDER BY order_index)
				) AS geom
			FROM players
			WHERE room_id = ?
		)
		SELECT
			l1.start_order AS first_start,
			l1.end_order AS first_end,
			l2.start_order AS second_start,
			l2.end_order AS second_end
		FROM lines l1, lines l2
		WHERE l1.start_order < l2.start_order
		  AND l1.geom IS NOT NULL
		  AND l2.geom IS NOT NULL
		  AND ST_Intersects(l1.geom, l2.geom)
		  AND NOT ST_Touches(l1.geom, l2.geom)
	`, roomID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get intersecting edge pairs: %w", err)
	}

	pairs := make([]IntersectingEdgePair, 0, len(rows))
	for _, r := range rows {
		pairs = append(pairs, IntersectingEdgePair{
			First: Edge{
				StartOrderIndex: r.FirstStart,
				EndOrderIndex:   r.FirstEnd,
			},
			Second: Edge{
				StartOrderIndex: r.SecondStart,
				EndOrderIndex:   r.SecondEnd,
			},
		})
	}

	return pairs, nil
}

// GetRoomIntersectionCount は room 内の交差辺ペア数を返す。
func (r *Repository) GetRoomIntersectionCount(ctx context.Context, roomID uuid.UUID, tx *gorm.DB) (int, error) {
	pairs, err := r.GetIntersectingEdgePairs(ctx, roomID, tx)
	if err != nil {
		return 0, err
	}

	return len(pairs), nil
}
