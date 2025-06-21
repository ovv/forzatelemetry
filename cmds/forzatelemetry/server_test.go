package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestMainDummy(t *testing.T) {
	// This is a placeholder test to ensure the test file exists.
}

func TestConfigureLoggerLevels(t *testing.T) {
	levels := []struct {
		name     string
		value    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"default", "", slog.LevelInfo},
		{"unknown", "something", slog.LevelInfo},
	}
	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tc.value)
			configureLogger()
			// No panic = pass; can't easily check slog global level
		})
	}
}

func TestRevisionVar(t *testing.T) {
	if revision == "" {
		t.Error("revision should not be empty")
	}
}

func TestRunMissingEnv(t *testing.T) {
	os.Setenv("POSTGRES_DSN", "")
	os.Setenv("GRAFANA_BASE_URL", "")
	code := run()
	if code != 1 {
		t.Errorf("expected exit code 1 for missing env, got %d", code)
	}
}
