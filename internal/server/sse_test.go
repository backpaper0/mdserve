package server_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mdserve/internal/server"
	"mdserve/internal/sse"
)

// --- Task 6.2: SSEHandler ---

// nonFlushingWriter is an http.ResponseWriter that does NOT implement http.Flusher.
// This allows testing the 500 fallback path in SSEHandler.
type nonFlushingWriter struct {
	header http.Header
	body   *bytes.Buffer
	code   int
}

func newNonFlushingWriter() *nonFlushingWriter {
	return &nonFlushingWriter{header: make(http.Header), body: &bytes.Buffer{}, code: 200}
}

func (w *nonFlushingWriter) Header() http.Header         { return w.header }
func (w *nonFlushingWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
func (w *nonFlushingWriter) WriteHeader(code int)        { w.code = code }

func TestSSEHandler_Returns500WhenFlusherNotSupported(t *testing.T) {
	broker := sse.New()
	handler := server.NewSSEHandler(broker)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := newNonFlushingWriter()

	handler.ServeHTTP(w, req)

	if w.code != http.StatusInternalServerError {
		t.Errorf("expected 500 when Flusher not supported, got %d", w.code)
	}
}

func TestSSEHandler_SetsCorrectContentType(t *testing.T) {
	broker := sse.New()
	handler := server.NewSSEHandler(broker)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestSSEHandler_NoCacheHeaders(t *testing.T) {
	broker := sse.New()
	handler := server.NewSSEHandler(broker)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()

	cc := resp.Header.Get("Cache-Control")
	if cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}
}

func TestSSEHandler_SendsReloadEventOnBroadcast(t *testing.T) {
	broker := sse.New()
	handler := server.NewSSEHandler(broker)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Broadcastを少し遅らせて呼ぶ
	go func() {
		time.Sleep(30 * time.Millisecond)
		broker.Broadcast()
	}()

	// "data: reload" を受信できること
	eventReceived := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "reload") {
				eventReceived <- line
				return
			}
		}
	}()

	select {
	case line := <-eventReceived:
		if !strings.Contains(line, "reload") {
			t.Errorf("expected 'reload' in SSE event, got %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Error("did not receive SSE reload event within 2s")
	}
}

// --- SSEがサーバーに統合されていることの確認 ---

func TestServer_SSEEndpointAvailable(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{DocRoot: t.TempDir(), Port: port, NoWatch: false}
	s := server.New(cfg)

	go s.Start()
	t.Cleanup(func() { s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/events", port))
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /events, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestServer_SSEEndpointAvailableInNoWatchMode(t *testing.T) {
	// --no-watch のときでも /events エンドポイントは存在する（ブロードキャストは行われない）
	port := freePort(t)
	cfg := server.Config{DocRoot: t.TempDir(), Port: port, NoWatch: true}
	s := server.New(cfg)

	go s.Start()
	t.Cleanup(func() { s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/events", port))
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /events even in no-watch mode, got %d", resp.StatusCode)
	}
}
