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
	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

// TestConcurrentWriteWithMutex verifies that the mutex fix prevents panics.
func TestConcurrentWriteWithMutex(t *testing.T) {
	pongWait = 10 * time.Millisecond
	pingPeriod = (pongWait * 9) / 10

	defer func() {
		pingPeriod = defaultPingPeriod
		pongWait = defaultPongWait
	}()

	ws, testFile, cleanup := setupConcurrencyTest(t)
	defer cleanup()

	messageCount, panicCount := runConcurrentWriteTest(t, ws, testFile)
	assert.Equal(t, messageCount, 0) // XXX
	assert.Equal(t, panicCount, 0)
}

func setupConcurrencyTest(t *testing.T) (*websocket.Conn, *os.File, func()) {
	t.Helper()

	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	if err != nil {
		t.Fatalf("%v", err)
	}

	_, _ = testFile.WriteString("INITIAL.\n")
	dir := filepath.Dir(testFile.Name())

	watcher, err := watcher.Init(dir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	s := httptest.NewServer(wsHandler(watcher))

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

func runConcurrentWriteTest(t *testing.T, ws *websocket.Conn, testFile *os.File) (int32, int32) {
	t.Helper()

	var (
		wg           sync.WaitGroup
		panicCount   atomic.Int32
		messageCount atomic.Int32
	)

	stopTest := make(chan bool)

	numWriters := 200

	for i := range numWriters {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)

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
		t.Logf("test timeout. messages: %d", messageCount.Load())
	}

	return messageCount.Load(), panicCount.Load()
}

func readWebSocketMessages(
	t *testing.T,
	ws *websocket.Conn,
	stopReading <-chan bool,
	panicCount *atomic.Int32,
	messageCount *atomic.Int32,
) {
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

func handleReaderPanic(t *testing.T, panicCount *atomic.Int32) {
	t.Helper()

	if r := recover(); r != nil {
		errStr := fmt.Sprintf("%v", r)
		if !strings.Contains(errStr, "repeated read") {
			panicCount.Add(1)
			t.Errorf("unexpected panic in reader: %v", r)
		}
	}
}

func readSingleMessage(t *testing.T, ws *websocket.Conn, messageCount *atomic.Int32) {
	t.Helper()

	_ = ws.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	msgType, msg, _ := ws.ReadMessage()

	if msgType == websocket.TextMessage && string(msg) == expectedReloadMsg {
		messageCount.Add(1)
		t.Logf("reload message received")
	}
}
