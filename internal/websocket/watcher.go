package websocket

import (
	"regexp"
	"sync"
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
	r := regexp.MustCompile(ignorePattern)
	once := sync.Once{}
	for {
		select {
		case event := <-watcher.Events:
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if r.MatchString(event.Name) {
					utils.LogDebug("Debug [ignore]: %s", event.Name)
				} else {
					once.Do(func() {
						utils.LogInfo("Change detected in %s, refreshing", event.Name)
						reload <- true
						go func() {
							time.Sleep(lockTime)
							once = sync.Once{}
						}()
					})
				}
			}
		case err := <-watcher.Errors:
			errorChan <- err
		case <-done:
			return
		}
	}
}
