package server

import (
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

const ignorePattern = `\.swp$|~$|^\.DS_Store$|^4913$`
const lockTime = 100 * time.Millisecond

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
		case event := <-watcher.Events:
			if !m.TryLock() {
				break
			}
			go func() {
				defer m.Unlock()
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if r.MatchString(event.Name) {
						utils.LogDebug("Debug [ignore]: %s", event.Name)
					} else {
						utils.LogInfo("Change detected in %s, refreshing", event.Name)
						reload <- true
						time.Sleep(lockTime)
					}
				}
			}()
		case err := <-watcher.Errors:
			errorChan <- err
		case <-done:
			return
		}
	}
}
