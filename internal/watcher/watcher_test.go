package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	debounce100 := time.Millisecond * 100

	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		debounce time.Duration
		wantErr  bool
	}{
		{
			name: "watcher creation successfull",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				return tempDir
			},
			debounce: debounce100,
			wantErr:  false,
		},
		{
			name: "return error if dir not exists",
			setupDir: func(t *testing.T) string {
				return "/dir_not_exists"
			},

			debounce: debounce100,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			w, err := NewWatcher(dir, tt.debounce)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewWatcher() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			defer w.watcher.Close()
			if w.dir != dir {
				t.Errorf("expected dir %s, got %s", dir, w.dir)
			}

			if w.debounce != tt.debounce {
				t.Errorf("expected debounce %v, got %v", tt.debounce, w.debounce)
			}
		})
	}
}

func TestStart(t *testing.T) {
	debounce50 := time.Millisecond * 50

	type testStep func(t *testing.T, dir string)

	writeFile := func(filename string, content string) testStep {
		return func(t *testing.T, dir string) {
			err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644)
			if err != nil {
				t.Fatalf("failed to write file: %v", err)
			}
		}
	}
	sleep := func(d time.Duration) testStep {
		return func(t *testing.T, dir string) {
			time.Sleep(d)
		}
	}

	tests := []struct {
		name        string
		debounce    time.Duration
		steps       []testStep
		cancelEarly bool
		wantMsg     int
	}{
		{
			name:     "debounce works as intended",
			debounce: debounce50,
			wantMsg:  1,
			steps: []testStep{
				writeFile("index.html", "1"),
				writeFile("index.html", "2"),
				writeFile("index.html", "3"),
				writeFile("index.html", "4"),
				writeFile("index.html", "5"),
			},
		},
		{
			name:        "watcher and timer exit properly if context is cancelled",
			debounce:    debounce50,
			cancelEarly: true,
			wantMsg:     0,
			steps: []testStep{
				writeFile("index.html", "late input"),
			},
		},
		{
			name:     "frequent events prolongate timer (moving window)",
			debounce: debounce50,
			wantMsg:  1,
			steps: []testStep{
				writeFile("index.html", "click 1"),
				sleep(30 * time.Millisecond),
				writeFile("index.html", "click 2"),
				sleep(30 * time.Millisecond),
				writeFile("index.html", "click 3"),
			},
		},
		{
			name:     "independent events trigger multiple times",
			debounce: debounce50,
			wantMsg:  2,
			steps: []testStep{
				writeFile("index.html", "batch 1"),
				sleep(70 * time.Millisecond),
				writeFile("index.html", "batch 2"),
				sleep(70 * time.Millisecond),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			events := make(chan string, 10)
			watcher, err := NewWatcher(dir, tt.debounce)
			if err != nil {
				t.Fatalf("error creating watcher: %v", err)
			}

			workerDone := make(chan struct{})

			go func() {
				watcher.Start(ctx, func(path string) {
					events <- path
				})
				close(workerDone)
			}()

			if tt.cancelEarly {
				cancel()
				<-workerDone
			}

			for _, step := range tt.steps {
				step(t, dir)
			}

			gotMessages := 0
			timeout := time.After(250 * time.Millisecond)

		mainLoop:
			for {
				select {
				case <-events:
					gotMessages++
				case <-timeout:
					break mainLoop
				}
			}

			if gotMessages != tt.wantMsg {
				t.Errorf("expected %d messages, got %d", tt.wantMsg, gotMessages)
			}
		})
	}
}
