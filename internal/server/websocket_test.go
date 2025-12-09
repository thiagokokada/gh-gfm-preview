package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

const expectedReloadMsg = "reload"

func TestWriter(t *testing.T) {
	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	assert.Nil(t, err)

	defer os.Remove(testFile.Name())

	_, _ = testFile.WriteString("BEFORE.\n")
	dir := filepath.Dir(testFile.Name())

	watcher, err := watcher.Init(dir)
	assert.Nil(t, err)

	s := httptest.NewServer(wsHandler(watcher))

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, res, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Nil(t, err)

	<-time.After(50 * time.Millisecond) // XXX

	defer ws.Close()
	defer res.Body.Close()
	defer s.Close()

	_, err = testFile.WriteString("AFTER.\n")
	assert.Nil(t, err)

	_, p, err := ws.ReadMessage()
	assert.Nil(t, err)

	actual := string(p)

	assert.Equal(t, actual, expectedReloadMsg)
}

func TestConcurrentWrites(t *testing.T) {
	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	assert.Nil(t, err)

	defer os.Remove(testFile.Name())

	_, _ = testFile.WriteString("INITIAL.\n")
	dir := filepath.Dir(testFile.Name())

	watcher, err := watcher.Init(dir)
	assert.Nil(t, err)

	s := httptest.NewServer(wsHandler(watcher))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, res, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Nil(t, err)

	defer ws.Close()
	defer res.Body.Close()

	<-time.After(50 * time.Millisecond)

	errorChan := startConcurrentWrites(t, testFile, 10)

	messageCount := readReloadMessages(t, ws, errorChan)

	t.Logf("successfully received %d reload messages without panic", messageCount)
}

func startConcurrentWrites(t *testing.T, testFile *os.File, numWrites int) <-chan error {
	t.Helper()

	var wg sync.WaitGroup

	errorChan := make(chan error, 20)

	for range numWrites {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := testFile.WriteString("WRITE.\n")
			if err != nil {
				errorChan <- err
			}

			time.Sleep(10 * time.Millisecond)
		}()
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	return errorChan
}

func readReloadMessages(t *testing.T, ws *websocket.Conn, errorChan <-chan error) int {
	t.Helper()

	messageCount := 0
	timeout := time.After(3 * time.Second)

	for {
		select {
		case err := <-errorChan:
			assert.Nil(t, err)
		case <-timeout:
			t.Logf("received %d messages before timeout", messageCount)

			return messageCount
		default:
			if tryReadReloadMessage(t, ws, &messageCount) {
				return messageCount
			}
		}
	}
}

func tryReadReloadMessage(t *testing.T, ws *websocket.Conn, messageCount *int) bool {
	t.Helper()

	err := ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	assert.Nil(t, err)

	msgType, msg, err := ws.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return true
		}

		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "i/o timeout") {
			t.Logf("read error (might be expected): %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		return false
	}

	if msgType == websocket.TextMessage && string(msg) == expectedReloadMsg {
		*messageCount++
		t.Logf("received reload message #%d", *messageCount)
	}

	return *messageCount >= 5
}

func TestConcurrentWritesStress(t *testing.T) {
	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	assert.Nil(t, err)

	defer os.Remove(testFile.Name())

	_, err = testFile.WriteString("INITIAL.\n")
	assert.Nil(t, err)

	dir := filepath.Dir(testFile.Name())

	watcher, err := watcher.Init(dir)
	assert.Nil(t, err)

	s := httptest.NewServer(wsHandler(watcher))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, res, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Nil(t, err)

	defer ws.Close()
	defer res.Body.Close()

	<-time.After(50 * time.Millisecond)

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, _ = testFile.WriteString("X")
		}()
	}

	done := make(chan bool)

	go func() {
		wg.Wait()

		done <- true
	}()

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-done:
			t.Log("all writes completed without panic")

			return
		case <-timeout:
			t.Log("test completed with timeout")

			return
		default:
			_ = ws.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			_, _, _ = ws.ReadMessage()

			time.Sleep(1 * time.Millisecond)
		}
	}
}
