package server

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
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
				utils.LogDebugf("[handshake error]: %v", err)
			}

			return
		}

		err = socket.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			utils.LogDebugf("[set read deadline error]: %v", err)
		}

		socket.SetPongHandler(func(string) error {
			err := socket.SetReadDeadline(time.Now().Add(pongWait))
			if err != nil {
				utils.LogDebugf("[set read deadline error in pong handler]: %v", err)
			}

			return nil
		})

		go wsReader(watcher.DoneCh, watcher.ErrorCh)
		go wsWriter(watcher.DoneCh, watcher.ErrorCh, watcher.ReloadCh)

		err = <-watcher.ErrorCh

		close(watcher.DoneCh)
		utils.LogInfof("Close WebSocket: %v\n", err)
		socket.Close()
	})
}

func wsReader(doneCh <-chan any, errorCh chan<- error) {
	for range doneCh {
		_, _, err := socket.ReadMessage()
		if err != nil {
			utils.LogDebugf("[read message]: %v", err)

			errorCh <- err
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
				utils.LogDebugf("[reload error]: %v", err)

				errorCh <- err
			}
		case <-ticker.C:
			utils.LogDebugf("[ping send]: ping to client")
			withSocketLock(func() {
				err = socket.WriteMessage(websocket.PingMessage, []byte{})
			})

			if err != nil {
				// Do nothing
				utils.LogDebugf("[ping error]: %v", err)
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
