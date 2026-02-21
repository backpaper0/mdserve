package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mdserve/internal/sse"
	"mdserve/internal/watcher"
)

// --- Task 6.1: FileWatcher ---

func TestWatcher_WatchStartsWithoutError(t *testing.T) {
	docRoot := t.TempDir()
	broker := sse.New()
	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}
}

func TestWatcher_BroadcastsOnFileWrite(t *testing.T) {
	docRoot := t.TempDir()

	// ブロードキャストを記録するブローカー
	recorded := make(chan struct{}, 10)
	broker := &recordingBroker{ch: recorded}

	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// ファイルを書き込む
	testFile := filepath.Join(docRoot, "test.md")
	if err := os.WriteFile(testFile, []byte("# Hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Broadcastが呼ばれること
	select {
	case <-recorded:
		// success
	case <-time.After(2 * time.Second):
		t.Error("Broadcast was not called after file write within 2s")
	}
}

func TestWatcher_BroadcastsOnFileRemove(t *testing.T) {
	docRoot := t.TempDir()

	// 先にファイルを作っておく
	testFile := filepath.Join(docRoot, "remove.md")
	if err := os.WriteFile(testFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	recorded := make(chan struct{}, 10)
	broker := &recordingBroker{ch: recorded}

	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// 既存イベントをドレイン
	time.Sleep(50 * time.Millisecond)
	drainChannel(recorded)

	// ファイルを削除する
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	select {
	case <-recorded:
		// success
	case <-time.After(2 * time.Second):
		t.Error("Broadcast was not called after file remove within 2s")
	}
}

func TestWatcher_BroadcastsOnFileRename(t *testing.T) {
	docRoot := t.TempDir()

	testFile := filepath.Join(docRoot, "original.md")
	if err := os.WriteFile(testFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	recorded := make(chan struct{}, 10)
	broker := &recordingBroker{ch: recorded}

	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	drainChannel(recorded)

	renamed := filepath.Join(docRoot, "renamed.md")
	if err := os.Rename(testFile, renamed); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	select {
	case <-recorded:
		// success
	case <-time.After(2 * time.Second):
		t.Error("Broadcast was not called after file rename within 2s")
	}
}

func TestWatcher_WatchesSubdirRecursively(t *testing.T) {
	docRoot := t.TempDir()
	subDir := filepath.Join(docRoot, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	recorded := make(chan struct{}, 10)
	broker := &recordingBroker{ch: recorded}

	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// サブディレクトリ内のファイル変更も検知すること
	testFile := filepath.Join(subDir, "sub.md")
	if err := os.WriteFile(testFile, []byte("# Sub"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-recorded:
		// success
	case <-time.After(2 * time.Second):
		t.Error("Broadcast was not called for subdirectory file write within 2s")
	}
}

func TestWatcher_DynamicallyWatchesNewSubdir(t *testing.T) {
	docRoot := t.TempDir()

	recorded := make(chan struct{}, 10)
	broker := &recordingBroker{ch: recorded}

	w := watcher.New(broker)
	defer func() { _ = w.Close() }()

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// 新しいサブディレクトリを作成
	newDir := filepath.Join(docRoot, "newdir")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// 少し待ってから新ディレクトリ内にファイルを作成
	time.Sleep(100 * time.Millisecond)
	drainChannel(recorded)

	testFile := filepath.Join(newDir, "new.md")
	if err := os.WriteFile(testFile, []byte("new content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-recorded:
		// success
	case <-time.After(3 * time.Second):
		t.Error("Broadcast was not called for file in newly created subdirectory within 3s")
	}
}

func TestWatcher_CloseStopsWatching(t *testing.T) {
	docRoot := t.TempDir()
	broker := sse.New()
	w := watcher.New(broker)

	if err := w.Watch(docRoot); err != nil {
		t.Fatalf("Watch() error: %v", err)
	}

	// Close は error なしに返ること
	if err := w.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

// --- helpers ---

// recordingBroker は Broadcast() 呼び出しをチャンネルに記録する。
type recordingBroker struct {
	ch chan struct{}
}

func (b *recordingBroker) Register() <-chan struct{}    { return make(chan struct{}) }
func (b *recordingBroker) Unregister(_ <-chan struct{}) {}
func (b *recordingBroker) Broadcast()                   { b.ch <- struct{}{} }

// drainChannel はチャンネルを空にする。
func drainChannel(ch chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
