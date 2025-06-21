package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"forzatelemetry/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"

	"github.com/jackc/pgx/v5/pgxpool"
	pgx "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *bun.DB
}

func NewStore(db *bun.DB) *Store {
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv("BUNDEBUG"),
	))

	return &Store{db}
}

func NewPGStore(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return NewStore(bun.NewDB(pgx.OpenDBFromPool(pool), pgdialect.New())), nil
}

func NewSqliteStore(dsn string) (*Store, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return nil, err
	}

	sqldb.SetMaxIdleConns(1000)
	sqldb.SetConnMaxLifetime(0)

	return NewStore(bun.NewDB(sqldb, sqlitedialect.New())), nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) IsPG() bool {
	return s.db.Dialect().Name() == dialect.PG
}

func (s *Store) IsSqlite() bool {
	return s.db.Dialect().Name() == dialect.SQLite
}

func (s *Store) Migrate(migrations *migrate.Migrations) error {
	if migrations == nil {
		slog.Info("no database migrations")
		return nil
	}
	slog.Info("starting database migrations")
	migrator := migrate.NewMigrator(s.db, migrations)

	err := migrator.Init(context.Background())
	if err != nil {
		return err
	}

	_, err = migrator.Migrate(context.Background())
	slog.Info("database migrations complete")
	return err
}

func (s *Store) CreateTables(ctx context.Context) error {
	_, err := s.db.NewCreateTable().IfNotExists().Model((*models.Point)(nil)).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.NewCreateTable().IfNotExists().Model((*models.Race)(nil)).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.NewCreateTable().IfNotExists().Model((*models.Track)(nil)).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.NewCreateTable().IfNotExists().Model((*models.Car)(nil)).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.NewCreateTable().IfNotExists().Model((*models.CarClass)(nil)).Exec(ctx)
	return err
}

func (s *Store) CreateIndexes(ctx context.Context) error {
	createIndex := s.db.NewCreateIndex().IfNotExists().Model((*models.Point)(nil)).Index("point_race_createdAt").Column("race", "created_at")

	if s.IsPG() {
		createIndex = createIndex.Include("lap_number")
	}
	_, err := createIndex.Exec(ctx)
	return err
}

func (s *Store) LoadFixtures(ctx context.Context, fixtures ...string) error {
	s.db.RegisterModel((*models.Race)(nil))
	s.db.RegisterModel((*models.Point)(nil))

	_, filename, _, _ := runtime.Caller(0)
	fixtureDir := os.DirFS(filepath.Join(filepath.Dir(filename), "../testutils/fixtures"))
	return dbfixture.New(s.db).Load(ctx, fixtureDir, fixtures...)
}

func (s *Store) Cleanup(ctx context.Context) error {
	if s.IsPG() {
		delRace := s.db.NewDelete().Model(&models.Race{}).Where("finished_at < CURRENT_DATE - interval '3 month'").Returning("id")
		_, err := s.db.NewDelete().With("delRace", delRace).Model(&models.Point{}).Where("race in (SELECT id FROM ?)", bun.Ident("delRace")).Exec(ctx)
		return err
	} else if s.IsSqlite() {
		// SQLite: delete points referencing old races, then delete old races
		_, err := s.db.NewDelete().Model(&models.Point{}).Where(
			"race IN (SELECT id FROM races WHERE finished_at < date('now', '-3 months'))",
		).Exec(ctx)
		if err != nil {
			return err
		}
		_, err = s.db.NewDelete().Model(&models.Race{}).Where(
			"finished_at < date('now', '-3 months')",
		).Exec(ctx)
		return err
	} else {
		return errors.New("unsupported database dialect")
	}
}

func (s *Store) UpsertTracks(tracks []models.Track, ctx context.Context) error {
	_, err := s.db.NewInsert().Model(&tracks).On("CONFLICT (ordinal) DO UPDATE").Set("name = EXCLUDED.name").Set("layout = EXCLUDED.layout").Set("location = EXCLUDED.location").Set("length = EXCLUDED.length").Exec(ctx)
	return err
}

func (s *Store) GetTrack(id int, ctx context.Context) (models.Track, error) {
	var track models.Track
	err := s.db.NewSelect().Model(&track).Column("*").Where("ordinal = ?", id).Scan(ctx)
	return track, err
}

func (s *Store) UpsertCars(cars []models.Car, ctx context.Context) error {
	_, err := s.db.NewInsert().Model(&cars).On("CONFLICT (ordinal) DO UPDATE").Set("year = EXCLUDED.year").Set("make = EXCLUDED.make").Set("model = EXCLUDED.model").Exec(ctx)
	return err
}

func (s *Store) GetCar(id int, ctx context.Context) (models.Car, error) {
	var car models.Car
	err := s.db.NewSelect().Model(&car).Column("*").Where("ordinal = ?", id).Scan(ctx)
	return car, err
}

func (s *Store) UpsertCarClasses(classes []models.CarClass, ctx context.Context) error {
	_, err := s.db.NewInsert().Model(&classes).On("CONFLICT (id) DO UPDATE").Set("name = EXCLUDED.name").Set("pi_start = EXCLUDED.pi_start").Set("pi_end = EXCLUDED.pi_end").Set("color = EXCLUDED.color").Exec(ctx)
	return err
}

func (s *Store) GetCarClass(id int, ctx context.Context) (models.CarClass, error) {
	var carClass models.CarClass
	err := s.db.NewSelect().Model(&carClass).Column("*").Where("id = ?", id).Scan(ctx)
	return carClass, err
}
