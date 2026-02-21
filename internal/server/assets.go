package server

import (
	"io/fs"
	"net/http"
)

// NewAssetHandler returns an http.Handler that serves static files from the given FS.
// Requests to /assets/foo.css are resolved to assets/foo.css in the provided FS.
func NewAssetHandler(assets fs.FS) http.Handler {
	return http.FileServer(http.FS(assets))
}
