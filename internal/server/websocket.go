package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

const (
	defaultPongWait   = 60 * time.Second
	defaultPingPeriod = (defaultPongWait * 9) / 10
)

var (
	upgrader   = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	pongWait   = defaultPongWait
	pingPeriod = defaultPingPeriod
	socket     *websocket.Conn
	socketMu   sync.Mutex
)

func wsHandler(watcher *watcher.Watcher) http.Handler {
	go watcher.Watch()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		socket, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			if errors.Is(err, websocket.HandshakeError{}) {
				slog.Error("Handshake error", "error", err)
			}

			return
		}
		defer socket.Close()

		err = socket.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			slog.Error("Set read deadline error", "error", err)
		}

		socket.SetPongHandler(func(string) error {
			err := socket.SetReadDeadline(time.Now().Add(pongWait))
			if err != nil {
				slog.Error("Set read deadline error in pong handler", "error", err)
			}

			return nil
		})

		go wsReader(watcher.DoneCh, watcher.ErrorCh)
		go wsWriter(watcher.DoneCh, watcher.ErrorCh, watcher.ReloadCh)

		err = <-watcher.ErrorCh

		if err != nil {
			slog.Error("Watcher channel error", "error", err)
		}

		close(watcher.DoneCh)
	})
}

func wsReader(doneCh <-chan any, errorCh chan<- error) {
	for range doneCh {
		_, _, err := socket.ReadMessage()
		if err != nil {
			slog.Error("Websocket reader read message error", "error", err)

			errorCh <- fmt.Errorf("websocket reader error: %w", err)
		}
	}
}

func wsWriter(doneCh <-chan any, errorCh chan<- error, reloadCh <-chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		var err error

		select {
		case <-reloadCh:
			withSocketLock(func() {
				err = socket.WriteMessage(websocket.TextMessage, []byte("reload"))
			})

			if err != nil {
				slog.Error("Websocket writer reload error", "error", err)

				errorCh <- fmt.Errorf("websocket writer error: %w", err)
			}
		case <-ticker.C:
			slog.Debug("Sending ping to client")
			withSocketLock(func() {
				err = socket.WriteMessage(websocket.PingMessage, []byte{})
			})

			if err != nil {
				// Do nothing
				slog.Debug("Ping error", "error", err)
			}
		case <-doneCh:
			return
		}
	}
}

func withSocketLock(fn func()) {
	socketMu.Lock()
	defer socketMu.Unlock()

	fn()
}
