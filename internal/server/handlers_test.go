package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdserve/internal/dirlist"
	"mdserve/internal/server"
	"mdserve/internal/tmpl"
)

// --- フェイク実装 ---

type fakeRenderer struct {
	content []byte
}

func (f *fakeRenderer) Render(_ string) ([]byte, error) {
	return f.content, nil
}

type fakeTmplEngine struct {
	lastPageData    *tmpl.PageData
	lastDirListData *tmpl.DirListData
}

func (f *fakeTmplEngine) RenderPage(data tmpl.PageData) ([]byte, error) {
	f.lastPageData = &data
	return []byte("<html>page</html>"), nil
}

func (f *fakeTmplEngine) RenderDirList(data tmpl.DirListData) ([]byte, error) {
	f.lastDirListData = &data
	return []byte("<html>dirlist</html>"), nil
}

// --- Task 4.1: directoryHandler ユニットテスト ---
// router経由でdirectoryHandlerをテストする（resolvedPathKeyが非公開なため）

func startTestServer(t *testing.T, docRoot string, tmplEngine tmpl.TemplateEngine, rend *fakeRenderer) *httptest.Server {
	t.Helper()
	dirHandler := server.NewDirectoryHandler(dirlist.New(), rend, tmplEngine, docRoot, false)
	router := server.NewRequestRouter(docRoot, http.NotFoundHandler(), dirHandler, http.NotFoundHandler())
	return httptest.NewServer(router)
}

// TestDirectoryHandler_ListParam_WithIndexFile は ?list あり + IndexFile あり のとき
// 一覧表示され、DirListData.IndexURL にパスがセットされることを検証する。
func TestDirectoryHandler_ListParam_WithIndexFile(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fakeTmpl := &fakeTmplEngine{}
	srv := startTestServer(t, docRoot, fakeTmpl, &fakeRenderer{content: []byte("<p>readme</p>")})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/?list")
	if err != nil {
		t.Fatalf("GET /?list: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if fakeTmpl.lastDirListData == nil {
		t.Fatal("expected RenderDirList to be called, but it was not")
	}
	if fakeTmpl.lastDirListData.IndexURL == "" {
		t.Errorf("expected IndexURL to be set when ?list and IndexFile exists, got empty")
	}
	if fakeTmpl.lastPageData != nil {
		t.Error("expected RenderPage NOT to be called when ?list is set")
	}
}

// TestDirectoryHandler_NoListParam_WithIndexFile は ?list なし + IndexFile あり のとき
// README を表示し、PageData.DirListURL にパスがセットされることを検証する。
func TestDirectoryHandler_NoListParam_WithIndexFile(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fakeTmpl := &fakeTmplEngine{}
	srv := startTestServer(t, docRoot, fakeTmpl, &fakeRenderer{content: []byte("<p>readme</p>")})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if fakeTmpl.lastPageData == nil {
		t.Fatal("expected RenderPage to be called, but it was not")
	}
	if fakeTmpl.lastPageData.DirListURL == "" {
		t.Errorf("expected DirListURL to be set when IndexFile exists and no ?list, got empty")
	}
	if !strings.HasSuffix(fakeTmpl.lastPageData.DirListURL, "?list") {
		t.Errorf("expected DirListURL to end with '?list', got %q", fakeTmpl.lastPageData.DirListURL)
	}
	if fakeTmpl.lastDirListData != nil {
		t.Error("expected RenderDirList NOT to be called when ?list is absent and IndexFile exists")
	}
}

// TestDirectoryHandler_ListParam_WithoutIndexFile は ?list あり + IndexFile なし のとき
// 一覧表示され、DirListData.IndexURL が空文字であることを検証する。
func TestDirectoryHandler_ListParam_WithoutIndexFile(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fakeTmpl := &fakeTmplEngine{}
	srv := startTestServer(t, docRoot, fakeTmpl, &fakeRenderer{content: []byte("<p>notes</p>")})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/?list")
	if err != nil {
		t.Fatalf("GET /?list: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if fakeTmpl.lastDirListData == nil {
		t.Fatal("expected RenderDirList to be called, but it was not")
	}
	if fakeTmpl.lastDirListData.IndexURL != "" {
		t.Errorf("expected IndexURL to be empty when no IndexFile exists, got %q", fakeTmpl.lastDirListData.IndexURL)
	}
}

// TestDirectoryHandler_ListParamWithValue は ?list=anything でも一覧表示されることを検証する
// （キーの存在チェック、値は無視）。
func TestDirectoryHandler_ListParamWithValue(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(docRoot, "README.md"), []byte("# README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fakeTmpl := &fakeTmplEngine{}
	srv := startTestServer(t, docRoot, fakeTmpl, &fakeRenderer{content: []byte("<p>readme</p>")})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/?list=anything")
	if err != nil {
		t.Fatalf("GET /?list=anything: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if fakeTmpl.lastDirListData == nil {
		t.Fatal("expected RenderDirList to be called for ?list=anything, but it was not")
	}
	if fakeTmpl.lastPageData != nil {
		t.Error("expected RenderPage NOT to be called when ?list=anything")
	}
}
