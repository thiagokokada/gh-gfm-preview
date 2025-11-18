package server

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
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
	mu         sync.Mutex
)

func wsHandler(watcher *fsnotify.Watcher) http.Handler {
	reload := make(chan bool, 1)
	errorChan := make(chan error)
	done := make(chan any)

	go watch(done, errorChan, reload, watcher)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		socket, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			if errors.Is(err, websocket.HandshakeError{}) {
				utils.LogDebugf("Debug [handshake error]: %v", err)
			}

			return
		}

		err = socket.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			utils.LogDebugf("Debug [set read deadline error]: %v", err)
		}

		socket.SetPongHandler(func(string) error {
			err := socket.SetReadDeadline(time.Now().Add(pongWait))
			if err != nil {
				utils.LogDebugf("Debug [set read deadline error in pong handler]: %v", err)
			}

			return nil
		})

		go wsReader(done, errorChan)
		go wsWriter(done, errorChan, reload)

		err = <-errorChan

		close(done)
		utils.LogInfof("Close WebSocket: %v\n", err)
		socket.Close()
	})
}

func wsReader(done <-chan any, errorChan chan<- error) {
	for range done {
		_, _, err := socket.ReadMessage()
		if err != nil {
			utils.LogDebugf("Debug [read message]: %v", err)

			errorChan <- err
		}
	}
}

func wsWriter(done <-chan any, errChan chan<- error, reload <-chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		var err error

		select {
		case <-reload:
			withLock(func() {
				err = socket.WriteMessage(websocket.TextMessage, []byte("reload"))
			})

			if err != nil {
				utils.LogDebugf("Debug [reload error]: %v", err)

				errChan <- err
			}
		case <-ticker.C:
			utils.LogDebugf("Debug [ping send]: ping to client")
			withLock(func() {
				err = socket.WriteMessage(websocket.PingMessage, []byte{})
			})

			if err != nil {
				// Do nothing
				utils.LogDebugf("Debug [ping error]: %v", err)
			}
		case <-done:
			return
		}
	}
}

func withLock(fn func()) {
	mu.Lock()
	defer mu.Unlock()

	fn()
}
