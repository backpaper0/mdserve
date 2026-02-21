package tmpl_test

import (
	"html/template"
	"strings"
	"testing"

	"mdserve/internal/dirlist"
	"mdserve/internal/tmpl"
)

// --- Task 3.2: HTMLページテンプレートエンジン ---

func TestRenderPage_ValidHTML(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Test Page",
		Content: template.HTML("<p>Hello World</p>"),
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected <!DOCTYPE html> in output")
	}
	if !strings.HasSuffix(strings.TrimSpace(html), "</html>") {
		t.Errorf("expected output to end with </html>")
	}
}

func TestRenderPage_ContainsTitle(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "My Document",
		Content: template.HTML("<p>body</p>"),
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	if !strings.Contains(string(got), "My Document") {
		t.Errorf("expected title 'My Document' in output, got:\n%s", string(got))
	}
}

func TestRenderPage_ContainsContent(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Test",
		Content: template.HTML("<h2>Section</h2><p>Text here</p>"),
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "<h2>Section</h2>") {
		t.Errorf("expected HTML content in output, got:\n%s", html)
	}
	if !strings.Contains(html, "markdown-body") {
		t.Errorf("expected markdown-body class wrapping content, got:\n%s", html)
	}
}

func TestRenderPage_ContainsAssetLinks(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Test",
		Content: template.HTML("<p>body</p>"),
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "/assets/github-markdown.css") {
		t.Errorf("expected /assets/github-markdown.css link, got:\n%s", html)
	}
	if !strings.Contains(html, "/assets/highlight.css") {
		t.Errorf("expected /assets/highlight.css link, got:\n%s", html)
	}
	if !strings.Contains(html, "/assets/mermaid.min.js") {
		t.Errorf("expected /assets/mermaid.min.js script, got:\n%s", html)
	}
}

func TestRenderPage_MermaidInitScript(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Test",
		Content: template.HTML("<p>body</p>"),
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "mermaid.initialize") {
		t.Errorf("expected mermaid.initialize() call, got:\n%s", html)
	}
	if !strings.Contains(html, "startOnLoad") {
		t.Errorf("expected startOnLoad in mermaid init, got:\n%s", html)
	}
}

func TestRenderPage_LiveReloadEnabled(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:      "Test",
		Content:    template.HTML("<p>body</p>"),
		LiveReload: true,
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "/events") {
		t.Errorf("expected SSE /events script when LiveReload=true, got:\n%s", html)
	}
	if !strings.Contains(html, "EventSource") {
		t.Errorf("expected EventSource when LiveReload=true, got:\n%s", html)
	}
	if !strings.Contains(html, "location.reload") {
		t.Errorf("expected location.reload when LiveReload=true, got:\n%s", html)
	}
}

func TestRenderPage_LiveReloadDisabled(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:      "Test",
		Content:    template.HTML("<p>body</p>"),
		LiveReload: false,
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if strings.Contains(html, "EventSource") {
		t.Errorf("expected NO EventSource when LiveReload=false, got:\n%s", html)
	}
}

func TestRenderPage_Breadcrumbs(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Sub Page",
		Content: template.HTML("<p>content</p>"),
		Breadcrumbs: []dirlist.Breadcrumb{
			{Label: "Root", URL: "/"},
			{Label: "subdir", URL: "/subdir/"},
		},
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="/"`) {
		t.Errorf("expected root breadcrumb link, got:\n%s", html)
	}
	if !strings.Contains(html, "Root") {
		t.Errorf("expected Root label in breadcrumbs, got:\n%s", html)
	}
	if !strings.Contains(html, "subdir") {
		t.Errorf("expected subdir in breadcrumbs, got:\n%s", html)
	}
}

func TestRenderDirList_ValidHTML(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.DirListData{
		Title: "Test Directory",
		Breadcrumbs: []dirlist.Breadcrumb{
			{Label: "Root", URL: "/"},
		},
		Entries: []dirlist.Entry{
			{Name: "readme.md", Path: "/readme.md", IsDir: false},
			{Name: "subdir", Path: "/subdir/", IsDir: true},
		},
	}

	got, err := engine.RenderDirList(data)
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected <!DOCTYPE html> in output")
	}
	if !strings.HasSuffix(strings.TrimSpace(html), "</html>") {
		t.Errorf("expected output to end with </html>")
	}
}

func TestRenderDirList_ContainsEntries(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.DirListData{
		Title: "My Docs",
		Entries: []dirlist.Entry{
			{Name: "notes.md", Path: "/notes.md", IsDir: false},
			{Name: "docs", Path: "/docs/", IsDir: true},
		},
	}

	got, err := engine.RenderDirList(data)
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "notes.md") {
		t.Errorf("expected notes.md in output, got:\n%s", html)
	}
	if !strings.Contains(html, "docs") {
		t.Errorf("expected docs in output, got:\n%s", html)
	}
	if !strings.Contains(html, `href="/notes.md"`) {
		t.Errorf("expected link to /notes.md, got:\n%s", html)
	}
	if !strings.Contains(html, `href="/docs/"`) {
		t.Errorf("expected link to /docs/, got:\n%s", html)
	}
}

func TestRenderDirList_LiveReload(t *testing.T) {
	engine := tmpl.New()

	withLR, err := engine.RenderDirList(tmpl.DirListData{
		Title:      "Test",
		LiveReload: true,
	})
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}
	if !strings.Contains(string(withLR), "EventSource") {
		t.Errorf("expected EventSource when LiveReload=true")
	}

	withoutLR, err := engine.RenderDirList(tmpl.DirListData{
		Title:      "Test",
		LiveReload: false,
	})
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}
	if strings.Contains(string(withoutLR), "EventSource") {
		t.Errorf("expected NO EventSource when LiveReload=false")
	}
}

func TestRenderDirList_Breadcrumbs(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.DirListData{
		Title: "Deep Dir",
		Breadcrumbs: []dirlist.Breadcrumb{
			{Label: "Root", URL: "/"},
			{Label: "parent", URL: "/parent/"},
			{Label: "Deep Dir", URL: "/parent/deep/"},
		},
	}

	got, err := engine.RenderDirList(data)
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Root") {
		t.Errorf("expected Root breadcrumb, got:\n%s", html)
	}
	if !strings.Contains(html, "parent") {
		t.Errorf("expected parent breadcrumb, got:\n%s", html)
	}
}

// --- Task 1 + 3: ナビゲーションリンク (DirListURL / IndexURL) ---

func TestRenderPage_DirListURL_ShowsLink(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:      "Test",
		Content:    template.HTML("<p>body</p>"),
		DirListURL: "/subdir/?list",
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "/subdir/?list") {
		t.Errorf("expected DirListURL link in output, got:\n%s", html)
	}
	if !strings.Contains(html, "ファイル一覧を表示") {
		t.Errorf("expected link text 'ファイル一覧を表示' in output, got:\n%s", html)
	}
}

func TestRenderPage_DirListURL_Empty_HidesLink(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.PageData{
		Title:   "Test",
		Content: template.HTML("<p>body</p>"),
		// DirListURL is empty (zero value)
	}

	got, err := engine.RenderPage(data)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	html := string(got)
	if strings.Contains(html, "ファイル一覧を表示") {
		t.Errorf("expected NO dir-list link when DirListURL is empty, got:\n%s", html)
	}
}

func TestRenderDirList_IndexURL_ShowsLink(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.DirListData{
		Title:    "Test",
		IndexURL: "/subdir/",
	}

	got, err := engine.RenderDirList(data)
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "/subdir/") {
		t.Errorf("expected IndexURL link in output, got:\n%s", html)
	}
	if !strings.Contains(html, "README を表示") {
		t.Errorf("expected link text 'README を表示' in output, got:\n%s", html)
	}
}

func TestRenderDirList_IndexURL_Empty_HidesLink(t *testing.T) {
	engine := tmpl.New()
	data := tmpl.DirListData{
		Title: "Test",
		// IndexURL is empty (zero value)
	}

	got, err := engine.RenderDirList(data)
	if err != nil {
		t.Fatalf("RenderDirList error: %v", err)
	}

	html := string(got)
	if strings.Contains(html, "README を表示") {
		t.Errorf("expected NO readme link when IndexURL is empty, got:\n%s", html)
	}
}
