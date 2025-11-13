package server

import (
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
	watcherMutex     sync.RWMutex
	watchedDirs      = make(map[string]bool)
	globalWatcher    *fsnotify.Watcher
)

func createWatcher(dir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, fmt.Errorf("failed to create watcher: %w", err)
	}

	globalWatcher = watcher

	utils.LogInfof("Watching %s/ for changes", dir)

	err = addDirToWatcher(watcher, dir)
	if err != nil {
		return watcher, fmt.Errorf("failed to add dir %s to watcher: %w", dir, err)
	}

	return watcher, nil
}

// addDirToWatcher adds a directory to the watcher if not already watched
func addDirToWatcher(watcher *fsnotify.Watcher, dir string) error {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if watchedDirs[dir] {
		return nil // Already watching this directory
	}

	err := watcher.Add(dir)
	if err != nil {
		return err
	}

	watchedDirs[dir] = true
	utils.LogDebugf("Debug [watching directory]: %s", dir)

	return nil
}

// AddDirectoryToWatch adds a directory to the global watcher
func AddDirectoryToWatch(dir string) error {
	if globalWatcher == nil {
		return fmt.Errorf("watcher not initialized")
	}

	return addDirToWatcher(globalWatcher, dir)
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
