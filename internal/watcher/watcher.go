package watcher

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

const (
	ignorePattern = `\.swp$|~$|^\.DS_Store$|^4913$`
	lockTime      = 100 * time.Millisecond
)

var ErrWatcherNotInitialized = errors.New("watcher not initialized")

type Watcher struct {
	DoneCh   chan any
	ErrorCh  chan error
	ReloadCh chan bool

	watcher     *fsnotify.Watcher
	watchedDirs sync.Map
}

func Init(dir string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	utils.LogDebugf("[watcher created]")

	watcher := Watcher{
		DoneCh:   make(chan any),
		ErrorCh:  make(chan error),
		ReloadCh: make(chan bool, 1),
		watcher:  fsWatcher,
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

	utils.LogInfof("Watching %s for changes", dir)

	return err
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

			utils.LogDebugf("[event]: op=%s name=%s", event.Op, event.Name)

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if re.MatchString(event.Name) {
					utils.LogDebugf("[ignore]: %s", event.Name)

					continue
				}

				if !mu.TryLock() {
					utils.LogDebugf("[event ignored]: op=%s name=%s", event.Op, event.Name)

					continue
				}

				go func() {
					defer mu.Unlock()

					utils.LogInfof("Change detected in %s, refreshing", event.Name)

					w.ReloadCh <- true

					time.Sleep(lockTime)
				}()
			}
		case err := <-w.watcher.Errors:
			w.ErrorCh <- err
		case <-w.DoneCh:
			return
		}
	}
}
