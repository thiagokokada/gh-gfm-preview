package server

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

var (
	watcherMu     sync.RWMutex
	globalWatcher *fsnotify.Watcher

	ErrWatcherNotInitialized = errors.New("watcher not initialized")
)

func createWatcher(dir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, fmt.Errorf("failed to create watcher: %w", err)
	}

	globalWatcher = watcher

	utils.LogInfof("Watching %s/ for changes", dir)

	err = addDirectoryToWatcher(dir)
	if err != nil {
		return watcher, fmt.Errorf("failed to add dir %s to watcher: %w", dir, err)
	}

	return watcher, nil
}

func addDirectoryToWatcher(dir string) error {
	if globalWatcher == nil {
		return ErrWatcherNotInitialized
	}

	watcherMu.Lock()
	defer watcherMu.Unlock()

	err := globalWatcher.Add(dir)
	if err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	utils.LogDebugf("Debug [watching directory]: %s", dir)

	return nil
}

func watch(
	done <-chan any,
	errorChan chan<- error,
	reload chan<- bool,
	watcher *fsnotify.Watcher,
) {
	r := regexp.MustCompile(ignorePattern)
	m := sync.Mutex{}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}

			utils.LogDebugf("Debug [event]: op=%s name=%s", event.Op, event.Name)

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if r.MatchString(event.Name) {
					utils.LogDebugf("Debug [ignore]: %s", event.Name)

					continue
				}

				if !m.TryLock() {
					utils.LogDebugf("Debug [event ignored]: op=%s name=%s", event.Op, event.Name)

					continue
				}

				go func() {
					defer m.Unlock()

					utils.LogInfof("Change detected in %s, refreshing", event.Name)

					reload <- true

					time.Sleep(lockTime)
				}()
			}
		case err := <-watcher.Errors:
			errorChan <- err
		case <-done:
			return
		}
	}
}
