package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"forzatelemetry/storage"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		store := storage.NewStore(db)
		err := store.CreateTables(ctx)
		if err != nil {
			return err
		}

		return store.CreateIndexes(ctx)
	}, func(ctx context.Context, db *bun.DB) error {
		return fmt.Errorf("no migration exist")
	})
}
