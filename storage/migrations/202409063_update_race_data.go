package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"

	"forzatelemetry/storage"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		store := storage.NewStore(db)

		offset := 0
		for {
			races, count, err := store.SelectRaces([]storage.Where{}, offset, ctx)
			if err != nil {
				return err
			}
			for _, race := range races {
				point, err := store.SelectLastPoint(race.ID.String(), ctx)
				if err != nil && errors.Is(err, sql.ErrNoRows) {
					continue
				} else if err != nil {
					return err
				}

				race.BestLap = point.BestLap
				race.Position = point.RacePosition
				race.RaceTime = point.CurrentRaceTime
				race.DistanceTraveled = point.DistanceTraveled
				store.UpsertRaces(ctx, race.Race)
			}

			offset = offset + len(races)
			if offset >= count {
				return nil
			}
		}

	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("no migration exist")
	})
}
