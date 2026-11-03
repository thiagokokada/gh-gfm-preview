package server

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

func TestReloadSurvivesConnectionErrors(t *testing.T) {
	w := watcher.NewDisabled()
	handler := wsHandler(w)
	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)
	t.Cleanup(func() { close(w.DoneCh) })

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	first := dialReloadClient(t, url)
	second := dialReloadClient(t, url)
	closed := dialReloadClient(t, url)
	assert.Nil(t, closed.Close())

	// A rejected upgrade must not panic or disturb existing connections.
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ws", nil))
	assert.Equal(t, recorder.Code, http.StatusBadRequest)

	_, response, err := websocket.DefaultDialer.Dial(url, http.Header{"Origin": {"https://example.com"}})
	if err == nil {
		t.Fatal("cross-origin handshake succeeded")
	}

	if response == nil {
		t.Fatal("cross-origin handshake returned no HTTP response")
	}
	defer response.Body.Close()

	assert.Equal(t, response.StatusCode, http.StatusForbidden)

	// An fsnotify error must not stop either client's writer.
	w.ErrorCh <- fsnotify.ErrEventOverflow

	// Repeat saves so this also covers a connection whose registration is
	// still finishing when its handshake returns.
	done := make(chan struct{})

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case w.MessageCh <- watcher.ReloadMessage:
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()

	t.Cleanup(func() {
		close(done)
		<-stopped
	})

	assertReloadMessage(t, first)
	assertReloadMessage(t, second)
}

func dialReloadClient(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	conn, response, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	t.Cleanup(func() { conn.Close() })

	return conn
}

func TestWriterClosesConnectionWhenQueueCloses(t *testing.T) {
	release := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)

			return
		}
		defer conn.Close()

		client := &wsClient{conn: conn, send: make(chan wsMessage)}
		close(client.send)
		client.writePump()
		// Keep the handler alive until the test has checked that the writer
		// itself closed the socket.
		<-release
	}))
	// Release the handler before closing the server.
	defer s.Close()
	defer close(release)

	conn := dialReloadClient(t, "ws"+strings.TrimPrefix(s.URL, "http"))
	assert.Nil(t, conn.SetReadDeadline(time.Now().Add(time.Second)))

	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("writer left the connection open")
	}

	if timeout, ok := errors.AsType[net.Error](err); ok && timeout.Timeout() {
		t.Fatal("writer did not close the connection before the deadline")
	}
}
