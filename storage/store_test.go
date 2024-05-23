package storage_test

import (
	"context"
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/storage"
	"forzatelemetry/storage/migrations"
	"forzatelemetry/testutils"
)

func TestMigrate(t *testing.T) {
	store, err := storage.NewSqliteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	defer store.Close()

	err = store.Migrate(nil)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	err = store.Migrate(migrations.Migrations)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestCreateTrable(t *testing.T) {
	store, err := storage.NewSqliteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	defer store.Close()

	err = store.CreateTables(context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	err = store.CreateTables(context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestCreateIndexes(t *testing.T) {
	store, err := storage.NewSqliteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	defer store.Close()

	err = store.CreateTables(context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	err = store.CreateIndexes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	err = store.CreateIndexes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestUpsertTracks(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	err := store.UpsertTracks(models.Tracks, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	track, err := store.GetTrack(0, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	tracks := []models.Track{{Ordinal: track.Ordinal, Location: "Foo"}}
	err = store.UpsertTracks(tracks, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	track, err = store.GetTrack(0, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if track.Location != "Foo" {
		t.Errorf("expected %s got %s", "Foo", track.Location)
	}
}

func TestUpsertCars(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	err := store.UpsertCars(models.Cars, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	car, err := store.GetCar(247, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	cars := []models.Car{{Ordinal: car.Ordinal, Make: "Foo"}}
	err = store.UpsertCars(cars, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	car, err = store.GetCar(247, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if car.Make != "Foo" {
		t.Errorf("expected %s got %s", "Foo", car.Make)
	}
}

func TestUpsertCarClasses(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	err := store.UpsertCarClasses(models.CarClasses, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	carClass, err := store.GetCarClass(0, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	carClasses := []models.CarClass{{Id: carClass.Id, Name: "Foo"}}
	err = store.UpsertCarClasses(carClasses, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	carClass, err = store.GetCarClass(0, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if carClass.Name != "Foo" {
		t.Errorf("expected %s got %s", "Foo", carClass.Name)
	}
}
