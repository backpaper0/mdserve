// Package dirlist provides directory listing functionality.
package dirlist

import "errors"

// Entry represents a file or directory in a listing.
type Entry struct {
	Name  string // Display name (file name)
	Path  string // URL path relative to docRoot
	IsDir bool   // true if this entry is a subdirectory
}

// Breadcrumb represents one element in a navigation breadcrumb trail.
type Breadcrumb struct {
	Label string // Display label
	URL   string // Link target URL
}

// Listing holds the directory contents for a given directory path.
type Listing struct {
	Title       string       // Page title (directory name)
	Breadcrumbs []Breadcrumb // From root to current directory
	Entries     []Entry      // .md files and subdirectories only
	IndexFile   string       // Path of README.md / index.md, or "" if none
}

// ErrForbidden is returned when dirPath is outside docRoot.
var ErrForbidden = errors.New("path is outside document root")

// DirectoryLister lists directory contents relative to docRoot.
type DirectoryLister interface {
	List(dirPath, docRoot string) (*Listing, error)
}
