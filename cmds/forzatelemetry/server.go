package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"forzatelemetry/models"
	"forzatelemetry/storage"
	"forzatelemetry/storage/migrations"
	"forzatelemetry/telemetry"
	fthttp "forzatelemetry/web"
)

var revision = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "undefined"
}()

func main() {
	os.Exit(run())
}

func run() int {

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		slog.Warn("missing configuration: POSTGRES_DSN")
		return 1
	}
	telemetryAddr := os.Getenv("TELEMETRY_ADDR")
	if telemetryAddr == "" {
		telemetryAddr = ":8000"
	}
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8000"
	}
	dashboardBaseUrl := os.Getenv("GRAFANA_BASE_URL")
	if dashboardBaseUrl == "" {
		slog.Warn("missing configuration: GRAFANA_BASE_URL")
		return 1
	}

	db, err := storage.NewPGStore(dsn)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		return 1
	}
	err = db.Migrate(migrations.Migrations)
	if err != nil {
		slog.Error("failed to migrate database", "error", err)
		return 1
	}
	err = db.UpsertTracks(models.Tracks, context.Background())
	if err != nil {
		slog.Error("failed to sync tracks", "error", err)
		return 1
	}
	err = db.UpsertCars(models.Cars, context.Background())
	if err != nil {
		slog.Error("failed to sync cars", "error", err)
		return 1
	}
	err = db.UpsertCarClasses(models.CarClasses, context.Background())
	if err != nil {
		slog.Error("failed to sync car classes", "error", err)
		return 1
	}

	var wg sync.WaitGroup
	errorC := make(chan bool, 2)

	telemetryServer := telemetry.NewServer(telemetryAddr, db, 5*time.Second)

	router := fthttp.Router(db, revision, dashboardBaseUrl)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	wg.Add(1)
	go func() {
		slog.Info("starting telemetry server", "addr", telemetryAddr)
		err := telemetryServer.ListenAndProcess()
		if errors.Is(err, telemetry.ErrServerClosed) {
			slog.Info("telemetry server closed")
		} else if err != nil {
			slog.Error("failed to start telemetry server", "error", err)
			errorC <- true
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		slog.Info("starting http server", "addr", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("http server closed")
		} else if err != nil {
			slog.Error("failed to start http server", "error", err)
			errorC <- true
		}
		wg.Done()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	exitcode := 0
	select {
	case s := <-stop:
		slog.Info("received signal, shutting down", "signal", s)
	case <-errorC:
		exitcode = 1
		slog.Info("startup failed, shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting http server shutdown")
		err := httpServer.Shutdown(ctx)
		if err != nil {
			slog.Error("failed to shutdown http server", "error", err)
		}
		slog.Info("http server shutdown complete")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting telemetry server shutdown")
		err := telemetryServer.Shutdown(ctx)
		if err != nil {
			slog.Error("failed to shutdown telemetry server", "error", err)
		}
		slog.Info("telemetry server shutdown complete")
	}()

	wg.Wait()
	return exitcode
}
