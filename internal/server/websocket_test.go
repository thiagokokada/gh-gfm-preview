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

var expectedReloadMsg = string(watcher.ReloadMessage)

func TestWriter(t *testing.T) {
	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	assert.Nil(t, err)

	defer os.Remove(testFile.Name())

	_, _ = testFile.WriteString("BEFORE.\n")
	dir := filepath.Dir(testFile.Name())

	w, err := watcher.Init(dir)
	assert.Nil(t, err)

	s := httptest.NewServer(wsHandler(w))

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, res, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Nil(t, err)

	<-time.After(50 * time.Millisecond) // XXX

	defer ws.Close()
	defer res.Body.Close()
	defer s.Close()

	_, err = testFile.WriteString("AFTER.\n")
	assert.Nil(t, err)

	_, actual, err := ws.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, string(actual), expectedReloadMsg)
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

	errorChan := startConcurrentWrites(t, testFile, 100)

	waitForConcurrentWrites(t, errorChan)
	assertReloadMessage(t, ws)
}

func startConcurrentWrites(t *testing.T, testFile *os.File, numWrites int) <-chan error {
	t.Helper()

	var wg sync.WaitGroup

	errorChan := make(chan error, 20)

	for range numWrites {
		wg.Go(func() {
			_, err := testFile.WriteString("WRITE.\n")
			if err != nil {
				errorChan <- err
			}

			time.Sleep(10 * time.Millisecond)
		})
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	return errorChan
}

func waitForConcurrentWrites(t *testing.T, errorChan <-chan error) {
	t.Helper()

	timeout := time.After(3 * time.Second)

	for {
		select {
		case err, ok := <-errorChan:
			if !ok {
				return
			}

			assert.Nil(t, err)
		case <-timeout:
			t.Fatal("timeout waiting for concurrent writes")
		}
	}
}

func assertReloadMessage(t *testing.T, ws *websocket.Conn) {
	t.Helper()

	err := ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	assert.Nil(t, err)

	msgType, msg, err := ws.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, msgType, websocket.TextMessage)
	assert.Equal(t, string(msg), expectedReloadMsg)
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
		wg.Go(func() {
			_, _ = testFile.WriteString("X")
		})
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
			t.Log("all writes completed without error")

			return
		case <-timeout:
			t.Fatal("test completed with timeout")

			return
		default:
			_ = ws.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			_, _, _ = ws.ReadMessage()

			time.Sleep(1 * time.Millisecond)
		}
	}
}
