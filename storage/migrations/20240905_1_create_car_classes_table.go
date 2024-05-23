package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"forzatelemetry/storage"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return storage.NewStore(db).CreateTables(ctx)
	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("no migration exist")
	})
}
