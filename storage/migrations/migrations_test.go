package migrations_test

import (
	"testing"

	"forzatelemetry/storage"
	"forzatelemetry/storage/migrations"
	"forzatelemetry/testutils"
)

func TestMigrateNewStore(t *testing.T) {
	store, err := storage.NewSqliteStore("file::memory:?cache=shared")
	defer store.Close()

	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	err = store.Migrate(migrations.Migrations)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	err = store.Migrate(migrations.Migrations)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestMigrateExistingStore(t *testing.T) {
	store := testutils.NewStore("races.yaml", "points.yaml")
	defer store.Close()

	err := store.Migrate(migrations.Migrations)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	err = store.Migrate(migrations.Migrations)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}
