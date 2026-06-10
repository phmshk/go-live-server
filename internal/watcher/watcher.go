package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher to provide recursive directory monitoring
type Watcher struct {
	watcher  *fsnotify.Watcher // system file event watcher
	dir      string            // directory being monitored
	mu       sync.Mutex        // mutex to protect concurrent access to the debounce timer
	timer    *time.Timer       // timer used to wait for file events to settle
	debounce time.Duration     // duration of the debounce
}

// NewWatcher initializes and returns Watcher instance.
// after creattion performs a recursive walk through the target directory
// to register all existing subfolders for monitoring.
func NewWatcher(dir string, debounce time.Duration) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:  watcher,
		dir:      dir,
		debounce: debounce,
	}

	// recursive walk
	err = w.watchFolders()
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return w, nil
}

// watchFolders walks the directory tree starting from w.dir and adds
// every discovered subdirectory to the underlying fsnotify monitoring pool.
func (w *Watcher) watchFolders() error {
	return filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return w.watcher.Add(path)
		}
		return nil
	})
}

// Start initiates a blocking event loop that processes filesystem notifications.
// When an eligible event occurs, debounce mechanism delays the onChange callback
// until the stream of events stops for the duration specified by debounce
func (w *Watcher) Start(ctx context.Context, onChange func(string)) {
	defer func() {
		w.watcher.Close()
		w.mu.Lock()
		if w.timer != nil {
			w.timer.Stop()
		}
		w.mu.Unlock()
	}()

	for {
		select {

		// Handle Graceful Shutdown when the context is cancelled
		case <-ctx.Done():
			log.Println("Stopping file watcher")
			return

			// Receive filesystem events from fsnotify channel
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			base := filepath.Base(event.Name)
			if isIgnoredFile(base) {
				continue
			}

			if eventIsCRUD(event) {

				// Dynamically register newly created subdirectories to the watcher
				if event.Has(fsnotify.Create) {
					fi, err := os.Stat(event.Name)
					if err != nil { // File might have been deleted instantly
						continue
					}
					if fi.IsDir() {
						w.watcher.Add(event.Name)
					}

				}

				// Capture the event name in a local variable to prevent race
				// inside the async time.AfterFunc when the loop goes on
				fileName := event.Name
				w.mu.Lock()
				// If previous timer is active, stop it since new event arrived
				if w.timer != nil {
					w.timer.Stop()
				}
				w.timer = time.AfterFunc(w.debounce, func() {
					log.Println("modified file:", event.Name)
					onChange(fileName)
				})
				w.mu.Unlock()
			}

			// Handle incoming system file watcher errors
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			log.Println("error:", err)
		}
	}
}

func eventIsCRUD(event fsnotify.Event) bool {
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
}

// isIgnoredFile returns true if the file is a hidden, temporary, or system file.
func isIgnoredFile(name string) bool {
	base := filepath.Base(name)

	// Skip current/empty directory references or hidden files
	if base == "." || strings.HasPrefix(base, ".") {
		return true
	}

	// Skip editor backup files
	if strings.HasSuffix(base, "~") {
		return true
	}

	// Skip specific extension types (.tmp, .swp)
	ext := filepath.Ext(base)
	if ext == ".tmp" || ext == ".swp" {
		return true
	}

	// Skip common temporary extensions and system files
	switch strings.ToLower(base) {
	case "thumbs.db", "desktop.ini":
		return true
	}

	return false
}
