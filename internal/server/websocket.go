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
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	testPongWait   time.Duration
	testPingPeriod time.Duration
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	socket     *websocket.Conn
	writeMutex sync.Mutex
)

func getPongWait() time.Duration {
	if testPongWait > 0 {
		return testPongWait
	}

	return pongWait
}

func getPingPeriod() time.Duration {
	if testPingPeriod > 0 {
		return testPingPeriod
	}

	return pingPeriod
}

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

		err = socket.SetReadDeadline(time.Now().Add(getPongWait()))
		if err != nil {
			utils.LogDebugf("Debug [set read deadline error]: %v", err)
		}

		socket.SetPongHandler(func(string) error {
			err := socket.SetReadDeadline(time.Now().Add(getPongWait()))
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
	ticker := time.NewTicker(getPingPeriod())
	defer ticker.Stop()

	for {
		select {
		case <-reload:
			writeMutex.Lock()

			err := socket.WriteMessage(websocket.TextMessage, []byte("reload"))

			writeMutex.Unlock()

			if err != nil {
				utils.LogDebugf("Debug [reload error]: %v", err)

				errChan <- err
			}
		case <-ticker.C:
			utils.LogDebugf("Debug [ping send]: ping to client")

			writeMutex.Lock()

			err := socket.WriteMessage(websocket.PingMessage, []byte{})

			writeMutex.Unlock()

			if err != nil {
				// Do nothing
				utils.LogDebugf("Debug [ping error]: %v", err)
			}
		case <-done:
			return
		}
	}
}
