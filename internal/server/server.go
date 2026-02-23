// Package server provides the HTTP server.
package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mdserve/internal/dirlist"
	"mdserve/internal/renderer"
	"mdserve/internal/sse"
	"mdserve/internal/tmpl"
	"mdserve/internal/watcher"
)

// Config holds application-wide settings derived from CLI flags.
type Config struct {
	DocRoot  string // Absolute path to the directory being served.
	Port     int    // TCP port to listen on (default: 3333).
	NoWatch  bool   // If true, disable file watching and live reload.
	AssetsFS fs.FS  // Embedded static assets FS; nil means no asset serving.
}

// Server manages the lifecycle of the HTTP server.
type Server struct {
	config  Config
	httpSrv *http.Server
	broker  sse.Broker
	watcher watcher.Watcher
}

// New creates a Server configured by cfg.
func New(cfg Config) *Server {
	b := sse.New()
	w := watcher.New(b)
	return &Server{config: cfg, broker: b, watcher: w}
}

// Start builds the request mux, begins listening on the configured port,
// prints a startup message, and blocks until SIGINT/SIGTERM is received
// or Shutdown is called.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Build business-logic components.
	rnd := renderer.New()
	tmplEngine := tmpl.New()
	lister := dirlist.New()
	liveReload := !s.config.NoWatch

	// Wire route handlers.
	mdH := NewMarkdownHandler(rnd, tmplEngine, liveReload)
	dirH := NewDirectoryHandler(lister, rnd, tmplEngine, s.config.DocRoot, liveReload)
	staticH := NewStaticFileHandler(s.config.DocRoot)
	router := NewRequestRouter(s.config.DocRoot, mdH, dirH, staticH)

	// Mount asset handler if an FS was provided.
	if s.config.AssetsFS != nil {
		mux.Handle("/assets/", NewAssetHandler(s.config.AssetsFS))
	}

	// SSE endpoint is always available; file watcher only starts when NoWatch is false.
	mux.HandleFunc("/events", NewSSEHandler(s.broker))

	mux.Handle("/", router)

	// Start file watcher unless disabled.
	if !s.config.NoWatch {
		if err := s.watcher.Watch(s.config.DocRoot); err != nil {
			return fmt.Errorf("file watcher: %w", err)
		}
	}

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	fmt.Printf("Serving %s on http://localhost:%d\n", s.config.DocRoot, s.config.Port)

	serverDone := make(chan error, 1)
	go func() {
		err := s.httpSrv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serverDone <- err
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverDone:
		return err
	case <-quit:
		if err := s.Shutdown(); err != nil {
			return err
		}
		return <-serverDone
	}
}

// Shutdown gracefully stops the HTTP server with up to 5 seconds for
// active connections to complete.
// Sequence: watcher.Close → broker.Shutdown → http.Server.Shutdown(5s)
func (s *Server) Shutdown() error {
	if s.watcher != nil {
		_ = s.watcher.Close()
	}
	if s.broker != nil {
		s.broker.Shutdown()
	}
	if s.httpSrv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}
