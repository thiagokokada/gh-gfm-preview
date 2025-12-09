package watcher

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	ignorePattern = `\.swp$|~$|^\.DS_Store$|^4913$`
	lockTime      = 100 * time.Millisecond
)

var (
	ReloadMessage            = []byte("reload")
	ErrWatcherNotInitialized = errors.New("watcher not initialized")
)

type Watcher struct {
	DoneCh    chan struct{}
	ErrorCh   chan error
	MessageCh chan []byte

	watcher     *fsnotify.Watcher
	watchedDirs sync.Map
}

func Init(dir string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	slog.Debug("Watcher created")

	watcher := Watcher{
		DoneCh:    make(chan struct{}),
		ErrorCh:   make(chan error),
		MessageCh: make(chan []byte, 1),
		watcher:   fsWatcher,
	}

	err = watcher.AddDirectory(dir)
	if err != nil {
		return nil, err
	}

	return &watcher, nil
}

func (w *Watcher) Close() error {
	err := w.watcher.Close()
	if err != nil {
		return fmt.Errorf("error during watcher close call: %w", err)
	}

	return nil
}

func (w *Watcher) AddDirectory(dir string) error {
	if w == nil {
		return ErrWatcherNotInitialized
	}

	if _, loaded := w.watchedDirs.LoadOrStore(dir, true); loaded {
		return nil // Already watching this directory
	}

	err := w.watcher.Add(dir)
	if err != nil {
		// Roll back the map if Add failed
		w.watchedDirs.Delete(dir)

		return fmt.Errorf("failed to add dir %s to watcher: %w", dir, err)
	}

	slog.Info("Watching directory for changes", "dir", dir)

	return nil
}

func (w *Watcher) Watch() {
	re := regexp.MustCompile(ignorePattern)
	mu := sync.Mutex{}

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				continue
			}

			path := event.Name
			op := event.Op

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if re.MatchString(event.Name) {
					slog.Debug("FS event from ignored pattern", "op", op, "path", path)

					continue
				}

				if !mu.TryLock() {
					slog.Debug("FS event debounced", "op", op, "path", path)

					continue
				}

				slog.Debug("FS event", "op", op, "path", path)

				go func() {
					defer mu.Unlock()

					slog.Info("Change detected, refreshing", "path", event.Name)

					w.MessageCh <- ReloadMessage

					time.Sleep(lockTime)
				}()
			}
		case err := <-w.watcher.Errors:
			slog.Error("FS watcher error", "error", err)

			w.ErrorCh <- err
		case <-w.DoneCh:
			return
		}
	}
}
