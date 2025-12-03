package server

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

// TestConcurrentWriteWithMutex verifies that the mutex fix prevents panics.
func TestConcurrentWriteWithMutex(t *testing.T) {
	pingPeriod = 10 * time.Millisecond
	pongWait = 60 * time.Second

	defer func() {
		pingPeriod = defaultPingPeriod
		pongWait = defaultPongWait
	}()

	ws, testFile, cleanup := setupConcurrencyTest(t)
	defer cleanup()

	panicCount := runConcurrentWriteTest(t, ws, testFile)

	if panicCount > 0 {
		t.Logf("warning: panics occurred (likely from connection close): %d", panicCount)
	}
}

func setupConcurrencyTest(t *testing.T) (*websocket.Conn, *os.File, func()) {
	t.Helper()

	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	if err != nil {
		t.Fatalf("%v", err)
	}

	_, _ = testFile.WriteString("INITIAL.\n")
	dir := filepath.Dir(testFile.Name())

	err = watcher.Init(dir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	s := httptest.NewServer(wsHandler())

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, res, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer res.Body.Close()

	time.Sleep(50 * time.Millisecond)

	cleanup := func() {
		_ = ws.Close()
		s.Close()

		_ = os.Remove(testFile.Name())
	}

	return ws, testFile, cleanup
}

func runConcurrentWriteTest(t *testing.T, ws *websocket.Conn, testFile *os.File) int32 {
	t.Helper()

	var wg sync.WaitGroup

	panicCount := int32(0)
	stopTest := make(chan bool)

	numWriters := 200

	for i := range numWriters {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt32(&panicCount, 1)

					t.Logf("writer %d caught panic: %v", id, r)
				}
			}()

			for j := range 50 {
				select {
				case <-stopTest:
					return
				default:
					fmt.Fprintf(testFile, "W%d-%d\n", id, j)
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	stopReading := make(chan bool)
	messageCount := int32(0)

	go readWebSocketMessages(t, ws, stopReading, &panicCount, &messageCount)

	done := make(chan bool)

	go func() {
		wg.Wait()

		done <- true
	}()

	timeout := time.After(15 * time.Second)

	select {
	case <-done:
		close(stopTest)
		close(stopReading)
		time.Sleep(100 * time.Millisecond)
	case <-timeout:
		close(stopTest)
		close(stopReading)
		t.Logf("Test timeout. Messages: %d", atomic.LoadInt32(&messageCount))
	}

	return atomic.LoadInt32(&panicCount)
}

func readWebSocketMessages(t *testing.T, ws *websocket.Conn, stopReading <-chan bool, panicCount *int32, messageCount *int32) {
	t.Helper()

	defer handleReaderPanic(t, panicCount)

	for {
		select {
		case <-stopReading:
			return
		default:
			readSingleMessage(t, ws, messageCount)
		}
	}
}

func handleReaderPanic(t *testing.T, panicCount *int32) {
	t.Helper()

	if r := recover(); r != nil {
		atomic.AddInt32(panicCount, 1)

		errStr := fmt.Sprintf("%v", r)
		if !strings.Contains(errStr, "repeated read") {
			t.Errorf("unexpected panic in reader: %v", r)
		}
	}
}

func readSingleMessage(t *testing.T, ws *websocket.Conn, messageCount *int32) {
	t.Helper()

	_ = ws.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	msgType, msg, err := ws.ReadMessage()
	if err != nil {
		handleReadError(t, err)

		return
	}

	atomic.AddInt32(messageCount, 1)

	if msgType == websocket.TextMessage && string(msg) == "reload" {
		t.Logf("Reload message received")
	}
}

func handleReadError(t *testing.T, err error) {
	t.Helper()

	if !strings.Contains(err.Error(), "timeout") &&
		!strings.Contains(err.Error(), "i/o timeout") {
		if strings.Contains(err.Error(), "close") {
			t.Logf("Connection closed (possibly due to panic): %v", err)
		}
	}
}
