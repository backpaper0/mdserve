// Package mdserve_test contains end-to-end integration tests for the full server stack.
package mdserve_test

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mdserve "mdserve"
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

// --- ui-theme Task 3.2: /assets/theme.css 配信統合テスト ---

// TestE2E_ThemeCSS_Served は /assets/theme.css が HTTP 200 と text/css Content-Type で
// 配信され、タイポグラフィ定義を含むことを検証する（Task 1.1 の受け入れテスト）。
func TestE2E_ThemeCSS_Served(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{
		DocRoot:  t.TempDir(),
		Port:     port,
		NoWatch:  true,
		AssetsFS: mdserve.Assets,
	}
	s := server.New(cfg)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/assets/theme.css", port))
	if err != nil {
		t.Fatalf("GET /assets/theme.css: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}

	body := readBody(t, resp)
	checks := []string{
		"18px",
		"1.7",
		"0.875em",
		"markdown-body",
	}
	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Errorf("theme.css: expected %q in content, got:\n%s", want, body)
		}
	}
}

// --- Task 7.2: HTTPサーバーE2Eテスト（埋め込みアセット）---

// TestE2E_EmbeddedAssetsServed は実際の埋め込みアセット（go:embed）が
// HTTPサーバーから正しく配信されることを検証する。
func TestE2E_EmbeddedAssetsServed(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{
		DocRoot:  t.TempDir(),
		Port:     port,
		NoWatch:  true,
		AssetsFS: mdserve.Assets,
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	paths := []string{
		"/assets/mermaid.min.js",
		"/assets/github-markdown.css",
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", port, p))
			if err != nil {
				t.Fatalf("GET %s: %v", p, err)
			}
			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("GET %s: expected 200, got %d", p, resp.StatusCode)
			}
		})
	}
}

// TestE2E_MDFileReturns200HTML は実際のHTTPサーバーで .md ファイルへの
// GETリクエストが200 HTMLを返すことを検証する。
func TestE2E_MDFileReturns200HTML(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "hello.md"), []byte("# Hello World\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{
		DocRoot: docRoot,
		Port:    port,
		NoWatch: true,
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/hello.md", port))
	if err != nil {
		t.Fatalf("GET /hello.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

// TestE2E_NonExistentPathReturns404 は存在しないパスへのリクエストが
// HTTP 404 を返すことを検証する。
func TestE2E_NonExistentPathReturns404(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{
		DocRoot: t.TempDir(),
		Port:    port,
		NoWatch: true,
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/notfound.md", port))
	if err != nil {
		t.Fatalf("GET /notfound.md: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// TestE2E_IndexFilePriority はディレクトリへのリクエストで README.md / index.md が
// 優先表示されることを検証する。
func TestE2E_IndexFilePriority(t *testing.T) {
	t.Run("README.md優先", func(t *testing.T) {
		docRoot := t.TempDir()
		if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README Content\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		port := freePort(t)
		cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
		s := server.New(cfg)

		go func() { _ = s.Start() }()
		t.Cleanup(func() { _ = s.Shutdown() })

		waitForServer(t, port)

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
		if err != nil {
			t.Fatalf("GET /: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		buf := make([]byte, 4096)
		n, _ := resp.Body.Read(buf)
		if !strings.Contains(string(buf[:n]), "README Content") {
			t.Error("expected README.md content in directory index response")
		}
	})

	t.Run("index.md優先", func(t *testing.T) {
		docRoot := t.TempDir()
		if err := os.WriteFile(filepath.Join(docRoot, "index.md"), []byte("# Index Content\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		port := freePort(t)
		cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
		s := server.New(cfg)

		go func() { _ = s.Start() }()
		t.Cleanup(func() { _ = s.Shutdown() })

		waitForServer(t, port)

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
		if err != nil {
			t.Fatalf("GET /: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		buf := make([]byte, 4096)
		n, _ := resp.Body.Read(buf)
		if !strings.Contains(string(buf[:n]), "Index Content") {
			t.Error("expected index.md content in directory index response")
		}
	})
}

// TestE2E_PathTraversalBlocked はパストラバーサル攻撃が適切に拒否されることを検証する。
func TestE2E_PathTraversalBlocked(t *testing.T) {
	port := freePort(t)
	cfg := server.Config{
		DocRoot: t.TempDir(),
		Port:    port,
		NoWatch: true,
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/..%%2F..%%2Fetc%%2Fpasswd", port))
	if err != nil {
		t.Fatalf("GET path traversal: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("path traversal should not return 200")
	}
}

// --- Task 7.3: SSEライブリロードの統合テスト ---

// TestSSELiveReload_FileChangeTriggersReload は実際のHTTPサーバーで SSE 接続後に
// ファイルを変更すると "data: reload" イベントが受信されることを検証する。
func TestSSELiveReload_FileChangeTriggersReload(t *testing.T) {
	docRoot := t.TempDir()
	mdFile := filepath.Join(docRoot, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Initial\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{
		DocRoot: docRoot,
		Port:    port,
		NoWatch: false, // ファイル監視有効
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	// SSE 接続を確立する
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/events", port))
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/events: expected 200, got %d", resp.StatusCode)
	}

	// バックグラウンドで SSE イベントを監視する
	reloadReceived := make(chan struct{}, 1)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "reload") {
				select {
				case reloadReceived <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	// ファイル監視が準備完了するのを待つ
	time.Sleep(100 * time.Millisecond)

	// ファイルを更新する
	if err := os.WriteFile(mdFile, []byte("# Updated\n"), 0o644); err != nil {
		t.Fatalf("WriteFile update: %v", err)
	}

	// reload イベントが受信されることを確認する
	select {
	case <-reloadReceived:
		// success
	case <-time.After(3 * time.Second):
		t.Error("ファイル変更後3秒以内にSSEリロードイベントが受信されなかった")
	}
}

// --- directory-listing-with-readme 統合テスト ---

// readBody はレスポンスボディを最大 64KB 読み込む。
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	buf := make([]byte, 64*1024)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n])
}

// TestE2E_DirWithReadme_ShowsDirListLink は README.md があるディレクトリへのリクエストで
// ファイル一覧リンク（?list）が HTML に含まれることを検証する。
func TestE2E_DirWithReadme_ShowsDirListLink(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
	s := server.New(cfg)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := readBody(t, resp)
	if !strings.Contains(body, "?list") {
		t.Errorf("expected '?list' link in README page HTML, got:\n%s", body)
	}
	if !strings.Contains(body, "ファイル一覧を表示") {
		t.Errorf("expected 'ファイル一覧を表示' link text in README page HTML, got:\n%s", body)
	}
}

// TestE2E_DirWithReadme_ListParam_ShowsReadmeLink は ?list 付きリクエストで
// ファイル一覧と README へのリンクが HTML に含まれることを検証する。
func TestE2E_DirWithReadme_ListParam_ShowsReadmeLink(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
	s := server.New(cfg)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/?list", port))
	if err != nil {
		t.Fatalf("GET /?list: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := readBody(t, resp)
	if !strings.Contains(body, "README を表示") {
		t.Errorf("expected 'README を表示' link in dir-list page, got:\n%s", body)
	}
	// README.md 自身もエントリに含まれることを確認
	if !strings.Contains(body, "README.md") {
		t.Errorf("expected README.md in dir-list entries, got:\n%s", body)
	}
	// notes.md も含まれることを確認
	if !strings.Contains(body, "notes.md") {
		t.Errorf("expected notes.md in dir-list entries, got:\n%s", body)
	}
}

// TestE2E_DirWithoutReadme_ListParam_NoReadmeLink は README なしディレクトリへの ?list リクエストで
// README リンクが HTML に含まれないことを検証する。
func TestE2E_DirWithoutReadme_ListParam_NoReadmeLink(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
	s := server.New(cfg)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/?list", port))
	if err != nil {
		t.Fatalf("GET /?list: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := readBody(t, resp)
	if strings.Contains(body, "README を表示") {
		t.Errorf("expected NO 'README を表示' link when no README exists, got:\n%s", body)
	}
}

// TestE2E_DirWithReadme_ListParam_ShowsBreadcrumbs は ?list ページにブレッドクラムが
// 正しく表示されることを検証する。
func TestE2E_DirWithReadme_ListParam_ShowsBreadcrumbs(t *testing.T) {
	docRoot := t.TempDir()
	subDir := filepath.Join(docRoot, "docs")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "README.md"), []byte("# Docs README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{DocRoot: docRoot, Port: port, NoWatch: true}
	s := server.New(cfg)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/docs/?list", port))
	if err != nil {
		t.Fatalf("GET /docs/?list: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := readBody(t, resp)
	if !strings.Contains(body, "Root") {
		t.Errorf("expected 'Root' breadcrumb in ?list page, got:\n%s", body)
	}
	if !strings.Contains(body, "docs") {
		t.Errorf("expected 'docs' breadcrumb in ?list page, got:\n%s", body)
	}
}

// TestSSELiveReload_NoWatchModeNoSSEEvent は --no-watch モードで起動した場合に
// ファイル変更後も SSE イベントが発行されないことを検証する。
func TestSSELiveReload_NoWatchModeNoSSEEvent(t *testing.T) {
	docRoot := t.TempDir()
	mdFile := filepath.Join(docRoot, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Initial\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	port := freePort(t)
	cfg := server.Config{
		DocRoot: docRoot,
		Port:    port,
		NoWatch: true, // ファイル監視無効
	}
	s := server.New(cfg)

	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })

	waitForServer(t, port)

	// SSE 接続を確立する
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/events", port))
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// バックグラウンドで SSE イベントを監視する
	reloadReceived := make(chan struct{}, 1)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "reload") {
				select {
				case reloadReceived <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	// ファイルを更新する
	if err := os.WriteFile(mdFile, []byte("# Updated\n"), 0o644); err != nil {
		t.Fatalf("WriteFile update: %v", err)
	}

	// --no-watch モードでは reload イベントが発行されないことを確認する
	select {
	case <-reloadReceived:
		t.Error("--no-watchモードではSSEリロードイベントが発行されないはずだが受信した")
	case <-time.After(500 * time.Millisecond):
		// success: no reload event received
	}
}
