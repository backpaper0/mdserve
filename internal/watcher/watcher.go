// Package watcher provides file system change detection.
package watcher

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"mdserve/internal/sse"
)

// Watcher monitors a directory tree for file changes.
type Watcher interface {
	// Watch starts recursive monitoring of docRoot in a background goroutine.
	Watch(docRoot string) error
	// Close stops monitoring and releases resources.
	Close() error
}

type fsWatcher struct {
	broker  sse.Broker
	watcher *fsnotify.Watcher
}

// New creates a Watcher that notifies broker on file changes.
func New(broker sse.Broker) Watcher {
	return &fsWatcher{broker: broker}
}

func (w *fsWatcher) Watch(docRoot string) error {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = fw

	// docRoot 以下の全ディレクトリを再帰的に監視対象へ追加する。
	if err := filepath.WalkDir(docRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // アクセス不可ディレクトリはスキップ
		}
		if d.IsDir() {
			return fw.Add(path)
		}
		return nil
	}); err != nil {
		_ = fw.Close()
		return err
	}

	go w.run()
	return nil
}

func (w *fsWatcher) run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			switch {
			case event.Has(fsnotify.Chmod):
				// Chmod イベントは無視する
			case event.Has(fsnotify.Create):
				// 新規ディレクトリの場合は動的に監視対象へ追加する
				if isDirPath(event.Name) {
					if err := w.watcher.Add(event.Name); err != nil {
						log.Printf("watcher: failed to add new dir %s: %v", event.Name, err)
					}
				}
				w.broker.Broadcast()
			default:
				// Write / Remove / Rename
				w.broker.Broadcast()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func (w *fsWatcher) Close() error {
	if w.watcher == nil {
		return nil
	}
	return w.watcher.Close()
}

// isDirPath reports whether path currently refers to a directory.
func isDirPath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
