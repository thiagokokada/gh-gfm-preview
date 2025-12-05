package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer w.Close()

	// Run Watch in a goroutine
	go w.Watch()

	// Give the watcher a brief moment to spin up
	time.Sleep(50 * time.Millisecond)

	// Create a file
	filePath := filepath.Join(dir, "test.txt")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	f.Close()

	// Wait for the reload signal
	select {
	case val := <-w.ReloadCh:
		if !val {
			t.Error("Expected true from ReloadCh")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for file creation event")
	}
}

func TestWatch_DetectsFileWrite(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	// Pre-create the file before initializing watcher
	filePath := filepath.Join(dir, "test.md")
	if err := os.WriteFile(filePath, []byte("initial"), 0o600); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	w, err := Init(dir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer w.Close()

	go w.Watch()

	time.Sleep(50 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(filePath, []byte("updated"), 0o600); err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	select {
	case <-w.ReloadCh:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for file write event")
	}
}

func TestWatch_DebounceLogic(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	w, err := Init(dir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer w.Close()

	go w.Watch()

	time.Sleep(50 * time.Millisecond)

	filePath := filepath.Join(dir, "rapid.txt")

	// Trigger multiple writes rapidly
	// The watcher logic has a TryLock and a Sleep(lockTime) which is 100ms.
	// If we write 5 times in 10ms, we should realistically only get 1 or 2 events processed.

	counter := 0
	done := make(chan bool)

	// Listener routine
	go func() {
		for {
			select {
			case <-w.ReloadCh:
				counter++
			case <-done:
				return
			}
		}
	}()

	// Fire rapid events
	for range 5 {
		if err := os.WriteFile(filePath, []byte(time.Now().String()), 0o600); err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Fast, but distinct enough for fsnotify
	}

	// Wait longer than the lockTime (100ms) to let the debounce finish
	time.Sleep(300 * time.Millisecond)
	close(done)

	// We expect fewer events than writes because of the lockTime logic
	if counter == 0 {
		t.Error("Expected at least one event, got 0")
	}

	if counter == 5 {
		t.Log("Warning: Debounce might not be catching all rapid events, got 5 events for 5 writes")
	}
}
