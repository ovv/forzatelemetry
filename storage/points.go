package storage

import (
	"context"
	"database/sql"
	"iter"
	"log/slog"

	"forzatelemetry/models"
)

func (s *Store) SelectLastPoint(race string, ctx context.Context) (models.Point, error) {
	var point models.Point
	err := s.db.NewSelect().Model(&point).Where("race = ?", race).Order("created_at DESC").Limit(1).Scan(ctx)
	return point, err
}

func (s *Store) IterPoints(race string, where []Where, ctx context.Context) iter.Seq2[models.Point, error] {
	points := []models.Point{}

	query := s.db.NewSelect().Model(&points).Column("*").Where("race = ?", race)
	query = addWhere(query, where)

	err := query.Order("created_at ASC").Scan(ctx)

	return func(yield func(models.Point, error) bool) {
		if err != nil {
			yield(models.Point{}, err)
			return
		}

		if len(points) == 0 {
			yield(models.Point{}, sql.ErrNoRows)
			return
		}

		slog.Info("points query info", "len", len(points), "cap", cap(points))
		for _, point := range points {
			if !yield(point, nil) {
				return
			}
		}
	}
}

func (s *Store) SelectLaps(race string, ctx context.Context) (map[uint16]models.Lap, error) {
	var _laps []models.Lap

	maxQ := s.db.NewSelect().Model(&models.Point{}).ColumnExpr("MAX(created_at) as max").Column("lap_number", "race").Where("race = ?", race).Group("lap_number", "race") // Group by race as well to benefit from the points_race_createdat index on the join
	err := s.db.NewSelect().Model(&models.Point{}).Column("lap_number", "current_lap", "created_at", "race_position", "current_race_time").Join("INNER JOIN (?) AS m", maxQ).JoinOn("m.max = point.created_at").JoinOn("m.race = point.race").Scan(ctx, &_laps)

	laps := make(map[uint16]models.Lap)
	if len(_laps) == 0 {
		return laps, sql.ErrNoRows
	}

	for _, l := range _laps {
		laps[l.LapNumber] = l
	}
	return laps, err
}

func (s *Store) InsertPoints(points []models.Point, ctx context.Context) error {
	_, err := s.db.NewInsert().Model(&points).Exec(ctx)
	return err
}
