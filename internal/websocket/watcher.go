package websocket

import (
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

const ignorePattern = `\.swp$|~$|^\.DS_Store$|^4913$`
const lockTime = 100 * time.Millisecond

func CreateWatcher(dir string) (*fsnotify.Watcher, error) {
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
	isLocked := false
	for {
		select {
		case event := <-watcher.Events:
			if isLocked {
				break
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				r := regexp.MustCompile(ignorePattern)
				if r.MatchString(event.Name) {
					utils.LogDebug("Debug [ignore]: %s", event.Name)
				} else {
					utils.LogInfo("Change detected in %s, refreshing", event.Name)
					isLocked = true
					reload <- true
					timer := time.NewTimer(lockTime)
					go func() {
						<-timer.C
						isLocked = false
					}()
				}
			}
		case err := <-watcher.Errors:
			errorChan <- err
		case <-done:
			return
		}
	}
}
