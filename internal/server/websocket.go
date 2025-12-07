package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
)

type wsClient struct {
	broker *wsBroker
	conn   *websocket.Conn
	send   chan []byte
}

func (c *wsClient) cleanup(doneCh chan<- struct{}) {
	defer c.conn.Close()

	// signal done and unregister
	c.broker.unregister <- c

	// notify that this client has ended
	doneCh <- struct{}{}
}

func (c *wsClient) readPump(errorCh chan<- error, doneCh chan<- struct{}) {
	defer c.cleanup(doneCh)

	for {
		// we only care about errors from ReadMessage; payloads are ignored here
		if _, _, err := c.conn.ReadMessage(); err != nil {
			slog.Debug("Websocket reader read message error", "error", err)

			errorCh <- fmt.Errorf("websocket reader error: %w", err)

			return
		}
	}
}

func (c *wsClient) writePump(errorCh chan<- error) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				slog.Debug("Broker closed the channel")

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Warn("Websocket writer write message error", "error", err)

				errorCh <- fmt.Errorf("websocket writer error: %w", err)

				return
			}

		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// do nothing
				slog.Debug("Ping error", "error", err)

				return
			}
		}
	}
}

func wsHandler(watcher *watcher.Watcher) http.Handler {
	broker := newBroker()
	go broker.run()

	// forward watcher reload signals to the broker
	go func() {
		for range watcher.ReloadCh {
			// non-blocking send; if broker.broadcast is full we drop the message briefly
			select {
			case broker.broadcast <- []byte("reload"):
			default:
				slog.Debug("Broker broadcast is busy, dropping reload")
			}
		}
	}()

	go watcher.Watch()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if errors.Is(err, websocket.HandshakeError{}) {
				slog.Error("Handshake error", "error", err)
			}

			return
		}

		err = conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			slog.Warn("Set read deadline error", "error", err)
		}

		client := &wsClient{
			broker: broker,
			conn:   conn,
			send:   make(chan []byte, 4),
		}

		broker.register <- client

		// per-connection done channel to allow the handler to wait for client close
		doneCh := make(chan struct{}, 1)

		go client.writePump(watcher.ErrorCh)
		go client.readPump(watcher.ErrorCh, doneCh)

		// wait until watcher signals an error (from anywhere) or the client connection
		// is done. If watcher signals an error, we close the client and let the broker
		// handle unregistering.
		select {
		case err := <-watcher.ErrorCh:
			if err != nil {
				slog.Debug("Watcher channel error", "error", err)
			}

			client.conn.Close()
		case <-doneCh:
			// client has disconnected normally
		}
	})
}
