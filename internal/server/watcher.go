package server

import (
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

func createWatcher(dir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, err
	}

	utils.LogInfo("Watching %s/ for changes", dir)
	err = watcher.Add(dir)

	return watcher, err
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
				break
			}

			utils.LogDebug("Debug [event]: op=%s name=%s", event.Op, event.Name)

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if r.MatchString(event.Name) {
					utils.LogDebug("Debug [ignore]: %s", event.Name)

					break
				}

				if !m.TryLock() {
					utils.LogDebug("Debug [event ignored]: op=%s name=%s", event.Op, event.Name)

					break
				}

				go func() {
					defer m.Unlock()
					utils.LogInfo("Change detected in %s, refreshing", event.Name)
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
