package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	dir     string
}

func NewWatcher(dir string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher: watcher,
		dir:     dir,
	}

	err = w.watchFolders()
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return w, nil
}

func (w *Watcher) watchFolders() error {
	return filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			log.Printf("Watching folder: %s", path)
			return w.watcher.Add(path)
		}
		return nil
	})
}

func (w *Watcher) Start(ctx context.Context) {
	go func() {
		defer w.watcher.Close()
		for {
			select {

			case <-ctx.Done():
				log.Println("Stopping file watcher")
				return

			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name)[0] == '.' {
					continue
				}
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}

				log.Println("error:", err)
			}
		}
	}()
}
