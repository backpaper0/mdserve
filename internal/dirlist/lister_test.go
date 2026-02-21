package dirlist_test

import (
	"os"
	"path/filepath"
	"testing"

	"mdserve/internal/dirlist"
)

// setup creates a temporary directory tree for testing.
// Returns (docRoot, subdir) absolute paths.
func setup(t *testing.T) (docRoot, subDir string) {
	t.Helper()
	docRoot = t.TempDir()
	subDir = filepath.Join(docRoot, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("setup: mkdir subdir: %v", err)
	}
	return docRoot, subDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// --- Task 5.3: DirectoryLister テスト ---

func TestList_READMEmdIsIndexFile(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "README.md"), "# Root")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if listing.IndexFile == "" {
		t.Error("expected IndexFile to be set for README.md, got empty")
	}
}

func TestList_IndexmdFallback(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "index.md"), "# Index")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if listing.IndexFile == "" {
		t.Error("expected IndexFile to be set for index.md, got empty")
	}
}

func TestList_READMEtakesPriorityOverIndex(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "README.md"), "# README")
	writeFile(t, filepath.Join(docRoot, "index.md"), "# Index")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if listing.IndexFile == "" {
		t.Error("expected IndexFile to be set")
	}
	if filepath.Base(listing.IndexFile) != "README.md" {
		t.Errorf("expected README.md to take priority, got %s", filepath.Base(listing.IndexFile))
	}
}

func TestList_NoIndexFile(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "notes.md"), "# Notes")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if listing.IndexFile != "" {
		t.Errorf("expected IndexFile to be empty, got %s", listing.IndexFile)
	}
}

func TestList_ExcludesNonMDFiles(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "notes.md"), "# Notes")
	writeFile(t, filepath.Join(docRoot, "image.png"), "PNG data")
	writeFile(t, filepath.Join(docRoot, "config.json"), "{}")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, e := range listing.Entries {
		if e.Name == "image.png" || e.Name == "config.json" {
			t.Errorf("non-.md file %q should not appear in Entries", e.Name)
		}
	}
}

func TestList_IncludesMDFiles(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "notes.md"), "# Notes")
	writeFile(t, filepath.Join(docRoot, "guide.md"), "# Guide")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	names := make(map[string]bool)
	for _, e := range listing.Entries {
		names[e.Name] = true
	}
	if !names["notes.md"] {
		t.Error("expected notes.md in Entries")
	}
	if !names["guide.md"] {
		t.Error("expected guide.md in Entries")
	}
}

func TestList_IncludesSubdirectories(t *testing.T) {
	docRoot, _ := setup(t)
	// subdir already created by setup

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	found := false
	for _, e := range listing.Entries {
		if e.Name == "subdir" && e.IsDir {
			found = true
		}
	}
	if !found {
		t.Error("expected subdir in Entries with IsDir=true")
	}
}

func TestList_DirEntryPathEndsWithSlash(t *testing.T) {
	docRoot, _ := setup(t)

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, e := range listing.Entries {
		if e.IsDir {
			if len(e.Path) == 0 || e.Path[len(e.Path)-1] != '/' {
				t.Errorf("directory entry path should end with '/', got %q", e.Path)
			}
		}
	}
}

func TestList_RootBreadcrumb(t *testing.T) {
	docRoot, subDir := setup(t)

	lister := dirlist.New()
	listing, err := lister.List(subDir, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(listing.Breadcrumbs) == 0 {
		t.Fatal("expected at least one breadcrumb")
	}
	if listing.Breadcrumbs[0].URL != "/" {
		t.Errorf("first breadcrumb URL should be '/', got %q", listing.Breadcrumbs[0].URL)
	}
}

func TestList_BreadcrumbsIncludeSubdir(t *testing.T) {
	docRoot, subDir := setup(t)

	lister := dirlist.New()
	listing, err := lister.List(subDir, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(listing.Breadcrumbs) < 2 {
		t.Fatalf("expected >= 2 breadcrumbs for subdir, got %d", len(listing.Breadcrumbs))
	}
	// Second breadcrumb should be the subdir
	if listing.Breadcrumbs[1].Label != "subdir" {
		t.Errorf("second breadcrumb label = %q, want 'subdir'", listing.Breadcrumbs[1].Label)
	}
	if listing.Breadcrumbs[1].URL != "/subdir/" {
		t.Errorf("second breadcrumb URL = %q, want '/subdir/'", listing.Breadcrumbs[1].URL)
	}
}

func TestList_RootBreadcrumbsHasOnlyRoot(t *testing.T) {
	docRoot, _ := setup(t)

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(listing.Breadcrumbs) != 1 {
		t.Errorf("expected 1 breadcrumb for root, got %d: %v", len(listing.Breadcrumbs), listing.Breadcrumbs)
	}
}

func TestList_ForbiddenOutsideRoot(t *testing.T) {
	docRoot, _ := setup(t)
	// Try to list the parent of docRoot
	parent := filepath.Dir(docRoot)

	lister := dirlist.New()
	_, err := lister.List(parent, docRoot)
	if err != dirlist.ErrForbidden {
		t.Errorf("expected ErrForbidden for path outside docRoot, got %v", err)
	}
}

func TestList_MDEntryPathStartsWithSlash(t *testing.T) {
	docRoot, _ := setup(t)
	writeFile(t, filepath.Join(docRoot, "notes.md"), "# Notes")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, e := range listing.Entries {
		if !e.IsDir {
			if len(e.Path) == 0 || e.Path[0] != '/' {
				t.Errorf("file entry path should start with '/', got %q", e.Path)
			}
		}
	}
}

func TestList_SubdirMDEntryHasCorrectPath(t *testing.T) {
	docRoot, subDir := setup(t)
	writeFile(t, filepath.Join(subDir, "doc.md"), "# Doc")

	lister := dirlist.New()
	listing, err := lister.List(subDir, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	found := false
	for _, e := range listing.Entries {
		if e.Name == "doc.md" {
			found = true
			if e.Path != "/subdir/doc.md" {
				t.Errorf("expected path '/subdir/doc.md', got %q", e.Path)
			}
		}
	}
	if !found {
		t.Error("expected doc.md in Entries")
	}
}

func TestList_CaseInsensitiveIndexDetection(t *testing.T) {
	docRoot, _ := setup(t)
	// Use lowercase readme.md
	writeFile(t, filepath.Join(docRoot, "readme.md"), "# readme")

	lister := dirlist.New()
	listing, err := lister.List(docRoot, docRoot)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if listing.IndexFile == "" {
		t.Error("expected IndexFile to be set for readme.md (case-insensitive), got empty")
	}
}
