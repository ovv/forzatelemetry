package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"

	"forzatelemetry/models"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		for _, column := range [][]string{{"best_lap", "real"}, {"race_time", "real"}, {"position", "smallint"}, {"distance_traveled", "real"}} {
			query := db.NewAddColumn().Model((*models.Race)(nil))

			var err error
			if db.Dialect().Name() == dialect.SQLite {
				query = query.ColumnExpr("COLUMN ? ?", bun.Ident(column[0]), bun.Safe(column[1]))
				_, err = query.Exec(ctx)
				if err != nil && err.Error() == fmt.Sprintf("SQL logic error: duplicate column name: %s (1)", column[0]) {
					err = nil
				}
			} else {
				query = query.ColumnExpr("COLUMN IF NOT EXISTS ? ?", bun.Ident(column[0]), bun.Safe(column[1]))
				_, err = query.Exec(ctx)
			}

			if err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("no migration exist")
	})
}
