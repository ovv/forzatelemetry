package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forzatelemetry/models"

	"github.com/uptrace/bun"
)

func (s *Store) makeRaceDetailledQuery(query *bun.SelectQuery) *bun.SelectQuery {
	query = query.ColumnExpr("races.*")
	query = query.Relation("TrackMetadata").Relation("CarMetadata").Relation("CarClassMetadata")
	return query
}

func buildDashboardUrl(race models.APIRace, baseUrl string) string {
	if baseUrl == "" {
		return ""
	}

	url := fmt.Sprintf("%s&var-race=%s&from=%d", baseUrl, race.ID, race.StartedAt.UnixMilli())
	if (race.FinishedAt != time.Time{}) {
		url += fmt.Sprintf("&to=%d", race.FinishedAt.UnixMilli())
	}
	return url
}

func (s *Store) SelectRaces(where []Where, offset int, ctx context.Context, dashboardBaseUrl string) ([]models.APIRace, int, error) {
	races := make([]models.APIRace, 0, 25)
	query := s.makeRaceDetailledQuery(s.db.NewSelect().Model(&races))

	query = addWhere(query, where)

	count, err := query.Order("finished_at DESC").Limit(25).Offset(offset).ScanAndCount(ctx)

	for i := range races {
		races[i].Dashboard = buildDashboardUrl(races[i], dashboardBaseUrl)
	}

	return races, count, err
}

func (s *Store) SelectRace(id string, ctx context.Context, dashboardBaseUrl string) (models.APIRace, error) {
	var race models.APIRace
	query := s.makeRaceDetailledQuery(s.db.NewSelect().Model(&race))
	err := query.Where("races.id = ?", id).Scan(ctx)
	race.Dashboard = buildDashboardUrl(race, dashboardBaseUrl)
	return race, err
}

func (s *Store) SelectRaceLaps(id string, ctx context.Context, dashboardBaseUrl string) (models.APIRaceDetailled, error) {
	race, err := s.SelectRace(id, ctx, dashboardBaseUrl)
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
