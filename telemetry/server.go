package telemetry

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"forzatelemetry/models"
	"forzatelemetry/storage"
)

const MAX_PACKET_SIZE = 1024
const CHAN_BUFFER = 100

var ErrServerClosed = errors.New("telemetry: Server closed")
var ErrAlreadyRunning = errors.New("server already running")

type Server struct {
	m  sync.Mutex
	wg sync.WaitGroup

	addr string

	listeners sync.Map
	db        *storage.Store
	server    net.PacketConn

	sessionCheckpointInterval time.Duration

	running bool
}

func NewServer(addr string, db *storage.Store, sessionCheckpointInterval time.Duration) *Server {
	if addr == "" {
		addr = ":8000"
	}

	return &Server{
		addr:                      addr,
		db:                        db,
		sessionCheckpointInterval: sessionCheckpointInterval,
	}
}

func (s *Server) Running() bool {
	return s.running
}

func (s *Server) Addr() net.Addr {
	if !s.Running() {
		return nil
	}
	return s.server.LocalAddr()
}

func (s *Server) Listeners() []string {
	var listeners []string
	s.listeners.Range(func(k any, v any) bool {
		listeners = append(listeners, k.(string))
		return true
	})
	return listeners
}

func (s *Server) ListenAndProcess() error {
	err := s.listen()
	if err != nil {
		return err
	}
	s.read()
	return ErrServerClosed
}

func (s *Server) listen() error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.Running() {
		return ErrAlreadyRunning
	}

	var err error
	s.server, err = net.ListenPacket("udp", s.addr)
	s.running = true
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()

	if !s.Running() {
		return nil
	}
	s.running = false

	done := make(chan error)
	go func() {
		done <- s.shutdown()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) shutdown() error {
	err := s.server.Close()
	if err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	s.listeners.Range(func(k any, v any) bool {
		close(v.(chan models.TelemetryPoint))
		return true
	})

	s.wg.Wait()
	return nil
}

func (s *Server) read() {
	buf := make([]byte, MAX_PACKET_SIZE)
	for {
		if !s.running {
			return
		}
		buf = buf[:]
		s.readOne(buf)
	}
}

func (s *Server) readOne(buf []byte) {
	_, addr, err := s.server.ReadFrom(buf)
	if err != nil {
		slog.Debug("failed reading from udp", "error", err)
		return
	}

	var point models.TelemetryPoint
	binary.Read(bytes.NewReader(buf), binary.LittleEndian, &point)

	if point.TimestampMS == 0 {
		slog.Debug("discarding 0 timestamp point")
		return
	}

	key := addr.String()
	select {
	case s.findChannel(key) <- point:
		return
	default:
		slog.Warn("session channel full", "session", key)
	}
}

func (s *Server) findChannel(key string) chan models.TelemetryPoint {
	sessionC, loaded := s.listeners.Load(key)
	if !loaded {
		c := make(chan models.TelemetryPoint, SESSION_CHANNEL_SIZE)
		sessionC, loaded = s.listeners.LoadOrStore(key, c)
		if !loaded {
			go s.process(key, c)
		}
	}
	return sessionC.(chan models.TelemetryPoint)
}

func (s *Server) process(key string, c chan models.TelemetryPoint) {
	s.wg.Add(1)
	defer s.wg.Done()

	session := NewSession(s.db)
	defer func() {
		err := session.Close()
		if err != nil {
			slog.Error("failed closing session", "error", err, "session", session.ID)
		}
	}()

	received := false
	var ok bool
	var err error
	var point models.TelemetryPoint

	ticker := time.NewTicker(s.sessionCheckpointInterval)
	for {
		select {
		case point, ok = <-c:
			if !ok {
				s.listeners.Delete(key)
				return
			}
			received = true
			if err == nil {
				err = session.Add(point)
				if err != nil {
					slog.Error("failed processing point", "error", err, "session", session.ID)
				}
			}
		case <-ticker.C:
			if !received {
				s.listeners.Delete(key)
				close(c)
				return
			}
			err = session.Checkpoint()
			if err != nil {
				slog.Error("failed checkpointing", "error", err, "session", session.ID)
			}
			received = false
		}
	}
}
