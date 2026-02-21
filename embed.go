// Package mdserve provides embedded static assets for the mdserve binary.
package mdserve

import "embed"

// Assets contains the embedded static files under the assets/ directory.
//
//go:embed assets
var Assets embed.FS
