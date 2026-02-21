package dirlist

import (
	"os"
	"path/filepath"
	"strings"
)

type dirLister struct{}

// New returns a DirectoryLister that reads from the real file system.
func New() DirectoryLister {
	return &dirLister{}
}

// List returns the contents of dirPath relative to docRoot.
// Only .md files and subdirectories are included in Entries.
// Returns ErrForbidden if dirPath is outside docRoot.
func (l *dirLister) List(dirPath, docRoot string) (*Listing, error) {
	cleanDir := filepath.Clean(dirPath)
	cleanRoot := filepath.Clean(docRoot)

	if !isWithinRoot(cleanDir, cleanRoot) {
		return nil, ErrForbidden
	}

	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		return nil, err
	}

	// Compute relative path from docRoot for URL construction.
	relDir, _ := filepath.Rel(cleanRoot, cleanDir)
	if relDir == "." {
		relDir = ""
	}

	// Find index file: README.md (case-insensitive) takes priority over index.md.
	indexFile := findIndexFile(cleanDir, entries)

	// Build entries: only .md files and subdirectories.
	var items []Entry
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			urlPath := "/" + filepath.ToSlash(filepath.Join(relDir, name)) + "/"
			items = append(items, Entry{Name: name, Path: urlPath, IsDir: true})
		} else if hasMDExtension(name) {
			urlPath := "/" + filepath.ToSlash(filepath.Join(relDir, name))
			items = append(items, Entry{Name: name, Path: urlPath, IsDir: false})
		}
	}

	title := filepath.Base(cleanDir)
	if cleanDir == cleanRoot {
		title = filepath.Base(cleanRoot)
	}

	return &Listing{
		Title:       title,
		Breadcrumbs: buildBreadcrumbs(cleanDir, cleanRoot),
		Entries:     items,
		IndexFile:   indexFile,
	}, nil
}

// findIndexFile searches entries for README.md (case-insensitive) then index.md.
func findIndexFile(dirPath string, entries []os.DirEntry) string {
	// First pass: look for README.md (case-insensitive)
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), "readme.md") {
			return filepath.Join(dirPath, e.Name())
		}
	}
	// Second pass: look for index.md (case-insensitive)
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), "index.md") {
			return filepath.Join(dirPath, e.Name())
		}
	}
	return ""
}

// hasMDExtension returns true if name ends with .md (case-insensitive).
func hasMDExtension(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".md")
}

// buildBreadcrumbs creates a breadcrumb trail from docRoot to dirPath.
// The first element is always {Label: "Root", URL: "/"}.
func buildBreadcrumbs(dirPath, docRoot string) []Breadcrumb {
	crumbs := []Breadcrumb{{Label: "Root", URL: "/"}}

	relPath, err := filepath.Rel(docRoot, dirPath)
	if err != nil || relPath == "." {
		return crumbs
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	url := "/"
	for _, part := range parts {
		if part == "" {
			continue
		}
		url += part + "/"
		crumbs = append(crumbs, Breadcrumb{Label: part, URL: url})
	}
	return crumbs
}

// isWithinRoot returns true if path equals root or is a descendant of root.
func isWithinRoot(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
