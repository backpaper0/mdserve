package server_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"mdserve/internal/server"
)

// freePort finds an available TCP port for testing.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// waitForServer polls until the server responds or times out.
func waitForServer(t *testing.T, port int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server on port %d did not start within 3s", port)
}

// --- Task 4.1: Config 構造体 ---

func TestConfig_Fields(t *testing.T) {
	cfg := server.Config{
		DocRoot: "/tmp/docs",
		Port:    3333,
		NoWatch: true,
	}
	if cfg.DocRoot != "/tmp/docs" {
		t.Errorf("DocRoot = %q, want /tmp/docs", cfg.DocRoot)
	}
	if cfg.Port != 3333 {
		t.Errorf("Port = %d, want 3333", cfg.Port)
	}
	if !cfg.NoWatch {
		t.Errorf("NoWatch = false, want true")
	}
}

func TestNew_ReturnsNonNilServer(t *testing.T) {
	cfg := server.Config{DocRoot: t.TempDir(), Port: 3333}
	s := server.New(cfg)
	if s == nil {
		t.Fatal("New() returned nil, want non-nil *Server")
	}
}

// --- Task 4.2: HTTPサーバー起動とグレースフルシャットダウン ---

func TestServer_StartListensOnConfiguredPort(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{DocRoot: t.TempDir(), Port: port}
	s := server.New(cfg)

	startErr := make(chan error, 1)
	go func() { startErr <- s.Start() }()

	waitForServer(t, port)

	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case err := <-startErr:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}
}

func TestServer_RespondsToHTTPRequests(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{DocRoot: t.TempDir(), Port: port}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 0 {
		t.Error("expected valid HTTP status code")
	}
}

func TestServer_ShutdownGracefully(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{DocRoot: t.TempDir(), Port: port}
	s := server.New(cfg)

	startDone := make(chan error, 1)
	go func() { startDone <- s.Start() }()
	waitForServer(t, port)

	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// Start() は Shutdown() 後に返ること
	select {
	case err := <-startDone:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}

	// シャットダウン後は接続が拒否されること
	_, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
	if err == nil {
		t.Error("expected connection error after shutdown, got success")
	}
}

func TestServer_ShutdownBeforeStartIsNoop(t *testing.T) {
	cfg := server.Config{DocRoot: t.TempDir(), Port: 3333}
	s := server.New(cfg)
	// Start を呼ばずに Shutdown を呼んでもパニックしないこと
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown before Start: %v", err)
	}
}
