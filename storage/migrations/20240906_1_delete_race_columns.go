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
		if db.Dialect().Name() == dialect.SQLite {
			_, err := db.NewDropColumn().Model((*models.Race)(nil)).ColumnExpr("paused_at").Exec(ctx)
			if err != nil && err.Error() != "SQL logic error: no such column: \"paused_at\" (1)" {
				return err
			}
			return nil
		} else {
			_, err := db.NewDropColumn().Model((*models.Race)(nil)).ColumnExpr("IF EXISTS paused_at").Exec(ctx)
			return err
		}

	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("no migration exist")
	})
}
