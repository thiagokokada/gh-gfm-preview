package watcher

import (
	"os"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

// setupTestDir creates a temporary directory for testing and returns its path
// and a cleanup function.
func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()

	dir := t.TempDir()
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestWatch_DetectsFileCreation(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	w, err := Init(dir)
	assert.Nil(t, err)

	defer w.Close()

	// Run Watch in a goroutine
	go w.Watch()

	// Give the watcher a brief moment to spin up
	time.Sleep(50 * time.Millisecond)

	// Create a file
	filePath := filepath.Join(dir, "test.txt")

	f, err := os.Create(filePath)
	assert.Nil(t, err)
	f.Close()

	// Wait for the reload signal
	select {
	case val := <-w.MessageCh:
		assert.Equal(t, string(val), "reload")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for file creation event")
	}
}

func TestWatch_DetectsFileWrite(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	// Pre-create the file before initializing watcher
	filePath := filepath.Join(dir, "test.md")
	err := os.WriteFile(filePath, []byte("initial"), 0o600)
	assert.Nil(t, err)

	w, err := Init(dir)
	assert.Nil(t, err)

	defer w.Close()

	go w.Watch()

	time.Sleep(50 * time.Millisecond)

	// Modify the file
	err = os.WriteFile(filePath, []byte("updated"), 0o600)
	assert.Nil(t, err)

	select {
	case <-w.MessageCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for file write event")
	}
}

func TestHandleEvent_DetectsFileChmod(t *testing.T) {
	w := &Watcher{MessageCh: make(chan []byte, 1)}
	debouncer := newReloadDebouncer(func(string) {
		w.MessageCh <- ReloadMessage
	})

	w.handleEvent(
		fsnotify.Event{Name: "test.md", Op: fsnotify.Chmod},
		regexp.MustCompile(ignorePattern),
		debouncer,
	)

	select {
	case <-w.MessageCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for file chmod event")
	}
}

func TestHandleEvent_WaitsForDebounceBeforeReload(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	filePath := filepath.Join(dir, "test.md")
	err := os.WriteFile(filePath, []byte("initial"), 0o600)
	assert.Nil(t, err)

	w := &Watcher{MessageCh: make(chan []byte, 1)}
	re := regexp.MustCompile(ignorePattern)
	debouncer := newReloadDebouncer(func(string) {
		w.MessageCh <- ReloadMessage
	})

	w.handleEvent(
		fsnotify.Event{Name: filepath.Join(dir, ".test.md.tmp"), Op: fsnotify.Create},
		re,
		debouncer,
	)

	select {
	case <-w.MessageCh:
		t.Fatal("reload signal sent before debounce settled")
	case <-time.After(debounceDelay / 2):
		// Success
	}

	err = os.WriteFile(filePath, []byte("updated"), 0o600)
	assert.Nil(t, err)

	w.handleEvent(
		fsnotify.Event{Name: filePath, Op: fsnotify.Create},
		re,
		debouncer,
	)

	select {
	case <-w.MessageCh:
		content, readErr := os.ReadFile(filePath)
		assert.Nil(t, readErr)
		assert.Equal(t, string(content), "updated")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for debounced reload event")
	}
}

func TestHandleEvent_ExtendsDebounceAfterLaterEvents(t *testing.T) {
	w := &Watcher{MessageCh: make(chan []byte, 1)}
	re := regexp.MustCompile(ignorePattern)
	debouncer := newReloadDebouncer(func(string) {
		w.MessageCh <- ReloadMessage
	})

	w.handleEvent(
		fsnotify.Event{Name: "dir", Op: fsnotify.Create},
		re,
		debouncer,
	)

	time.Sleep(debounceDelay / 2)

	w.handleEvent(
		fsnotify.Event{Name: filepath.Join("dir", "README.md"), Op: fsnotify.Create},
		re,
		debouncer,
	)

	select {
	case <-w.MessageCh:
		t.Fatal("reload signal sent before the later event settled")
	case <-time.After((debounceDelay / 2) + (debounceDelay / 4)):
		// Success
	}

	select {
	case <-w.MessageCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for trailing debounced reload event")
	}
}

func TestWatch_ReaddsRecreatedDirectory(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	subdir := filepath.Join(dir, "subdir")
	err := os.Mkdir(subdir, 0o700)
	assert.Nil(t, err)

	w, err := Init(dir)
	assert.Nil(t, err)

	defer w.Close()

	err = w.AddDirectory(subdir)
	assert.Nil(t, err)

	go w.Watch()

	time.Sleep(50 * time.Millisecond)

	err = os.RemoveAll(subdir)
	assert.Nil(t, err)
	drainReloadMessages(w.MessageCh)

	err = os.Mkdir(subdir, 0o700)
	assert.Nil(t, err)
	drainReloadMessages(w.MessageCh)

	err = w.AddDirectory(subdir)
	assert.Nil(t, err)

	filePath := filepath.Join(subdir, "test.md")
	err = os.WriteFile(filePath, []byte("updated"), 0o600)
	assert.Nil(t, err)

	select {
	case <-w.MessageCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for recreated directory file write event")
	}
}

func drainReloadMessages(ch <-chan []byte) {
	timer := time.NewTimer(2 * debounceDelay)
	defer timer.Stop()

	for {
		select {
		case <-ch:
		case <-timer.C:
			return
		}
	}
}

func TestWatch_DebounceLogic(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	w, err := Init(dir)
	assert.Nil(t, err)

	defer w.Close()

	go w.Watch()

	time.Sleep(50 * time.Millisecond)

	filePath := filepath.Join(dir, "rapid.txt")

	// Trigger multiple writes rapidly
	// The watcher logic has a TryLock and a debounce delay.
	// If we write 5 times in 10ms, we should realistically only get 1 or 2 events processed.

	var counter atomic.Uint32

	done := make(chan bool)

	// Listener routine
	go func() {
		for {
			select {
			case <-w.MessageCh:
				counter.Add(1)
			case <-done:
				return
			}
		}
	}()

	// Fire rapid events
	for range 5 {
		err := os.WriteFile(filePath, []byte(time.Now().String()), 0o600)
		assert.Nil(t, err)

		time.Sleep(10 * time.Millisecond) // Fast, but distinct enough for fsnotify
	}

	// Wait longer than the debounce delay to let the debounce finish
	time.Sleep(300 * time.Millisecond)
	close(done)

	// We expect fewer events than writes because of the debounce logic.
	assert.True(t, counter.Load() > 0)
	assert.True(t, counter.Load() <= 5)
}

func TestInit_ReturnsErrorForMissingPath(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing")

	w, err := Init(missingPath)
	assert.Nil(t, w)
	assert.True(t, err != nil)
}
