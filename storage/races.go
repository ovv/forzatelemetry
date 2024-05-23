package storage

import (
	"context"
	"database/sql"
	"errors"

	"forzatelemetry/models"

	"github.com/uptrace/bun"
)

func (s *Store) makeRaceDetailledQuery(query *bun.SelectQuery) *bun.SelectQuery {
	query = query.ColumnExpr("races.*")

	if s.IsPG() {
		query = query.ColumnExpr("(? || '&var-race=' || races.id || '&from=' || FLOOR(EXTRACT(EPOCH FROM started_at) * 1000) || (CASE WHEN finished_at is NULL THEN '' ELSE ('&to=' || FLOOR(EXTRACT(EPOCH FROM finished_at) * 1000)) END)) as dashboard", s.dashboardBaseUrl)
	}

	query = query.Relation("TrackMetadata").Relation("CarMetadata").Relation("CarClassMetadata")
	return query
}

func (s *Store) SelectRaces(where []Where, offset int, ctx context.Context) ([]models.APIRace, int, error) {
	races := make([]models.APIRace, 0, 25)
	query := s.makeRaceDetailledQuery(s.db.NewSelect().Model(&races))

	query = addWhere(query, where)

	count, err := query.Order("finished_at DESC").Limit(25).Offset(offset).ScanAndCount(ctx)
	return races, count, err
}

func (s *Store) SelectRace(id string, ctx context.Context) (models.APIRace, error) {
	var race models.APIRace
	query := s.makeRaceDetailledQuery(s.db.NewSelect().Model(&race))
	err := query.Where("races.id = ?", id).Scan(ctx)
	return race, err
}

func (s *Store) SelectRaceLaps(id string, ctx context.Context) (models.APIRaceDetailled, error) {
	race, err := s.SelectRace(id, ctx)
	if err != nil {
		return models.APIRaceDetailled{}, err
	}

	laps, err := s.SelectLaps(id, ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.APIRaceDetailled{}, err
	}

	return models.MakeRaceDetailled(race, laps), nil
}

func (s *Store) UpsertRaces(ctx context.Context, races ...models.Race) error {
	_, err := s.db.NewInsert().Model(&races).On(
		"CONFLICT (id) DO UPDATE").Set("paused = EXCLUDED.paused").Set("in_progress = EXCLUDED.in_progress").Set("finished_at = EXCLUDED.finished_at").Set("best_lap = EXCLUDED.best_lap").Set("race_time = EXCLUDED.race_time").Set("position = EXCLUDED.position").Set("distance_traveled = EXCLUDED.distance_traveled").Exec(context.Background())
	return err
}
