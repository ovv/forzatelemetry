package telemetry_test

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"forzatelemetry/models"
	"forzatelemetry/telemetry"
	"forzatelemetry/testutils"
)

func TestServerNoAddr(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	server := telemetry.NewServer("", store, 5*time.Second)

	if server.Addr() != nil {
		t.Errorf("expected nil got %v", server.Addr())
	}
}

func TestServer(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	server := telemetry.NewServer("127.0.0.1:0", store, 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	cancel()

	if server.Running() == true {
		t.Errorf("expected false got %v", server.Running())
	}

	go func(t *testing.T) {
		err := server.ListenAndProcess()
		if err == nil {
			t.Error("expected error got nil")
		} else if !errors.Is(err, telemetry.ErrServerClosed) {
			t.Errorf("expected %v got %v", telemetry.ErrServerClosed, err)
		}
	}(t)

	// let server start
	addr := server.Addr()
	for addr == nil {
		time.Sleep(10 * time.Microsecond)
		addr = server.Addr()
	}

	if server.Running() == false {
		t.Errorf("expected true got %v", server.Running())
	}

	err = server.ListenAndProcess()
	if err == nil {
		t.Error("expected error got nil")
	} else if !errors.Is(err, telemetry.ErrAlreadyRunning) {
		t.Errorf("expected %v got %v", telemetry.ErrAlreadyRunning, err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	cancel()
}

func TestServerGarbageData(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	server := telemetry.NewServer("127.0.0.1:0", store, 5*time.Second)
	go func(t *testing.T) {
		err := server.ListenAndProcess()
		if err == nil {
			t.Error("expected error got nil")
		} else if !errors.Is(err, telemetry.ErrServerClosed) {
			t.Errorf("expected %v got %v", telemetry.ErrServerClosed, err)
		}
	}(t)

	addr := server.Addr()
	for addr == nil {
		time.Sleep(10 * time.Microsecond)
		addr = server.Addr()
	}

	con, err := net.Dial("udp", addr.String())
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	_, err = con.Write([]byte("hi"))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	time.Sleep(10 * time.Microsecond)

	listeners := server.Listeners()
	if len(listeners) != 0 {
		t.Errorf("expected 0 got %v", len(listeners))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	cancel()
}

func TestServerValidData(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	server := telemetry.NewServer("127.0.0.1:0", store, 5*time.Second)
	go func(t *testing.T) {
		err := server.ListenAndProcess()
		if err == nil {
			t.Error("expected error got nil")
		} else if !errors.Is(err, telemetry.ErrServerClosed) {
			t.Errorf("expected %v got %v", telemetry.ErrServerClosed, err)
		}
	}(t)

	addr := server.Addr()
	for addr == nil {
		time.Sleep(10 * time.Microsecond)
		addr = server.Addr()
	}

	// first connection
	con, err := net.Dial("udp", addr.String())
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	point := testutils.Point(
		testutils.ParseUUID("cfb8395b-1bd3-4723-a2de-eb192365865b"),
		time.Now(),
		1,
	)

	err = binary.Write(con, binary.LittleEndian, point.TelemetryPoint)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	listeners := server.Listeners()
	if len(listeners) != 1 {
		t.Fatalf("expected 1 got %v", len(listeners))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	races, count, err := store.SelectRaces(nil, 0, ctx)
	cancel()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 got %v", count)
	}
	if len(races) != 1 {
		t.Fatalf("expected 1 got %v", len(races))
	}
	if races[0].CarPerformanceIndex != 100 {
		t.Fatalf("expected 100 got %v", races[0].CarPerformanceIndex)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	for _, err := range store.IterPoints(races[0].ID.String(), nil, ctx) {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected %v got %v", sql.ErrNoRows, err)
		}
		break
	}
	cancel()

	// second connection
	con, err = net.Dial("udp", addr.String())
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	point = testutils.Point(
		testutils.ParseUUID("a4fff587-4394-46be-baed-da3c178862a9"),
		time.Now(),
		1,
	)

	err = binary.Write(con, binary.LittleEndian, point.TelemetryPoint)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	listeners = server.Listeners()
	if len(listeners) != 2 {
		t.Fatalf("expected 2 got %v", len(listeners))
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	races, count, err = store.SelectRaces(nil, 0, ctx)
	cancel()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 got %v", count)
	}
	if len(races) != 2 {
		t.Fatalf("expected 2 got %v", len(races))
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	cancel()

	listeners = server.Listeners()
	if len(listeners) != 0 {
		t.Fatalf("expected 0 got %v", len(listeners))
	}
}

func TestServerCheckpoint(t *testing.T) {
	store := testutils.NewStore()
	defer store.Close()

	checkpointInterval := 500 * time.Millisecond
	server := telemetry.NewServer("127.0.0.1:0", store, checkpointInterval)
	go func(t *testing.T) {
		err := server.ListenAndProcess()
		if err == nil {
			t.Error("expected error got nil")
		} else if !errors.Is(err, telemetry.ErrServerClosed) {
			t.Errorf("expected %v got %v", telemetry.ErrServerClosed, err)
		}
	}(t)

	addr := server.Addr()
	for addr == nil {
		time.Sleep(10 * time.Microsecond)
		addr = server.Addr()
	}

	con, err := net.Dial("udp", addr.String())
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	point := testutils.Point(
		testutils.ParseUUID("cfb8395b-1bd3-4723-a2de-eb192365865b"),
		time.Now(),
		1,
	)

	err = binary.Write(con, binary.LittleEndian, point.TelemetryPoint)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	time.Sleep(checkpointInterval / 10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	races, _, err := store.SelectRaces(nil, 0, ctx)
	cancel()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(races) != 1 {
		t.Fatalf("expected 1 got %v", len(races))
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	for _, err := range store.IterPoints(races[0].ID.String(), nil, ctx) {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected %v got %v", sql.ErrNoRows, err)
		}
		break
	}
	cancel()

	time.Sleep(checkpointInterval)

	var points []models.Point
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	for point, err := range store.IterPoints(races[0].ID.String(), nil, ctx) {
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		points = append(points, point)
	}
	cancel()
	if len(points) != 1 {
		t.Fatalf("expected 1 got %v", len(points))
	}

	listeners := server.Listeners()
	if len(listeners) != 1 {
		t.Fatalf("expected 1 got %v", len(listeners))
	}

	// don't send anypoint in a checkpoint interval. Session is closed
	time.Sleep(checkpointInterval)
	listeners = server.Listeners()
	if len(listeners) != 0 {
		t.Fatalf("expected 0 got %v", len(listeners))
	}
}
