package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWriter(t *testing.T) {
	testFile, err := os.CreateTemp(t.TempDir(), "markdown-preview-test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.Remove(testFile.Name())

	_, _ = testFile.WriteString("BEFORE.\n")
	dir := filepath.Dir(testFile.Name())

	watcher, err := createWatcher(dir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	s := httptest.NewServer(wsHandler(watcher))

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	<-time.After(50 * time.Millisecond) // XXX

	defer ws.Close()
	defer s.Close()

	_, err = testFile.WriteString("AFTER.\n")
	if err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actual := string(p)
	expected := "reload"

	if actual != expected {
		t.Errorf("got %v\n want %v", actual, expected)
	}
}
