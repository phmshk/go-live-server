package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	dir := t.TempDir()

	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	w.watcher.Close()

	if w.dir != dir {
		t.Errorf("expected dir %q, got %q", dir, w.dir)
	}
}

func TestNewWatcher_NonExistentDir(t *testing.T) {
	_, err := NewWatcher("/nonexistent-path-for-test")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestNewWatcher_WatchesSubdirs(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	w.watcher.Close()
}

func TestStart_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		w.Start(ctx, func(fileName string) {
			t.Error("onChange should not be called after cancellation")
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestStart_FileWriteEvent(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan string, 1)
	go w.Start(ctx, func(fileName string) {
		changed <- fileName
	})

	time.Sleep(200 * time.Millisecond)

	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case fileName := <-changed:
		if fileName != testFile {
			t.Errorf("expected %q, got %q", testFile, fileName)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for file change event")
	}
}

func TestStart_FileWriteInSubdir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan string, 1)
	go w.Start(ctx, func(fileName string) {
		changed <- fileName
	})

	time.Sleep(200 * time.Millisecond)

	testFile := filepath.Join(subdir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case fileName := <-changed:
		if fileName != testFile {
			t.Errorf("expected %q, got %q", testFile, fileName)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for file change event in subdirectory")
	}
}

func TestStart_HiddenFileSkipped(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan string, 1)
	go w.Start(ctx, func(fileName string) {
		changed <- fileName
	})

	time.Sleep(200 * time.Millisecond)

	hiddenFile := filepath.Join(dir, ".hidden")
	if err := os.WriteFile(hiddenFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case fileName := <-changed:
		t.Errorf("onChange should not be called for hidden files, got %q", fileName)
	case <-time.After(1 * time.Second):
	}
}

func TestStart_MultipleWriteEvents(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan string, 2)
	go w.Start(ctx, func(fileName string) {
		changed <- fileName
	})

	time.Sleep(200 * time.Millisecond)

	files := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "b.txt"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	received := make(map[string]bool)
	for i := 0; i < len(files); i++ {
		select {
		case fileName := <-changed:
			received[fileName] = true
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for event %d", i)
		}
	}

	for _, f := range files {
		if !received[f] {
			t.Errorf("missing event for %q", f)
		}
	}
}
