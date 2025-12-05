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

	w           *fsnotify.Watcher
	mu          sync.Mutex
	watchedDirs map[string]bool
}

func Init(dir string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	utils.LogDebugf("Debug [watcher created]")

	watcher := Watcher{
		DoneCh:   make(chan any),
		ErrorCh:  make(chan error),
		ReloadCh: make(chan bool, 1),
		w:        fsWatcher,

		watchedDirs: make(map[string]bool),
	}

	err = watcher.AddDirectory(dir)
	if err != nil {
		return nil, err
	}

	return &watcher, nil
}

func (watcher *Watcher) Close() error {
	err := watcher.w.Close()
	if err != nil {
		return fmt.Errorf("error during watcher close call: %w", err)
	}

	return nil
}

func (watcher *Watcher) AddDirectory(dir string) error {
	if watcher == nil {
		return ErrWatcherNotInitialized
	}

	if watcher.watchedDirs[dir] {
		return nil // Already watching this directory
	}

	err := watcher.addDirectory(dir)

	utils.LogInfof("Watching %s for changes", dir)

	return err
}

func (watcher *Watcher) addDirectory(dir string) error {
	watcher.mu.Lock()
	defer watcher.mu.Unlock()

	err := watcher.w.Add(dir)
	if err != nil {
		return fmt.Errorf("failed to add dir %s to watcher: %w", dir, err)
	}

	watcher.watchedDirs[dir] = true

	return nil
}

func (watcher *Watcher) Watch() {
	re := regexp.MustCompile(ignorePattern)
	mu := sync.Mutex{}

	for {
		select {
		case event, ok := <-watcher.w.Events:
			if !ok {
				continue
			}

			utils.LogDebugf("Debug [event]: op=%s name=%s", event.Op, event.Name)

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if re.MatchString(event.Name) {
					utils.LogDebugf("Debug [ignore]: %s", event.Name)

					continue
				}

				if !mu.TryLock() {
					utils.LogDebugf("Debug [event ignored]: op=%s name=%s", event.Op, event.Name)

					continue
				}

				go func() {
					defer mu.Unlock()

					utils.LogInfof("Change detected in %s, refreshing", event.Name)

					watcher.ReloadCh <- true

					time.Sleep(lockTime)
				}()
			}
		case err := <-watcher.w.Errors:
			watcher.ErrorCh <- err
		case <-watcher.DoneCh:
			return
		}
	}
}
