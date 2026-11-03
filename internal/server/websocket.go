package server

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thiagokokada/gh-gfm-preview/internal/watcher"
)

const (
	writeWait         = 10 * time.Second
	defaultPongWait   = 60 * time.Second
	defaultPingPeriod = (defaultPongWait * 9) / 10 // must be less than pong wait
)

var (
	upgrader   = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	pongWait   = defaultPongWait
	pingPeriod = defaultPingPeriod
)

type wsClient struct {
	broker *wsBroker
	conn   *websocket.Conn
	send   chan wsMessage
}

func (c *wsClient) cleanup(doneCh chan<- struct{}) {
	defer c.conn.Close()

	// signal done and unregister
	c.broker.unregister <- c

	// notify that this client has ended
	doneCh <- struct{}{}
}

func (c *wsClient) remoteAddr() net.Addr {
	return c.conn.UnderlyingConn().RemoteAddr()
}

func (c *wsClient) readPump(doneCh chan<- struct{}) {
	defer c.cleanup(doneCh)

	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		slog.Warn("WS set read deadline error", "remote_addr", c.remoteAddr(), "error", err)
	}

	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			slog.Warn("WS set read deadline error", "remote_addr", c.remoteAddr(), "error", err)
		}

		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			slog.Debug("WS read message error", "remote_addr", c.remoteAddr(), "error", err)

			return
		}
	}
}

func (c *wsClient) writePump() {
	defer c.conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				slog.Debug("Broker closed the channel", "remote_addr", c.remoteAddr())

				return
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Debug("WS set write deadline error", "remote_addr", c.remoteAddr(), "error", err)

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg.message); err != nil {
				slog.Debug("WS write message error", "remote_addr", c.remoteAddr(), "error", err)

				return
			}

		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				slog.Debug("WS ping error", "remote_addr", c.remoteAddr(), "error", err)

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
		for {
			select {
			case message := <-watcher.MessageCh:
				broker.broadcast <- wsMessage{message: message}
			case err := <-watcher.ErrorCh:
				// Watcher errors are logged by Watch; keep clients connected.
				slog.Debug("Continuing live reload after watcher error", "error", err)
			case <-watcher.DoneCh:
				return
			}
		}
	}()

	go watcher.Watch()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Debug("WS connection upgrade error", "remote_addr", r.RemoteAddr, "error", err)

			return
		}

		client := &wsClient{
			broker: broker,
			conn:   conn,
			send:   make(chan wsMessage, 4),
		}

		broker.register <- client

		// per-connection done channel to allow the handler to wait for client close
		doneCh := make(chan struct{}, 1)

		go client.writePump()
		go client.readPump(doneCh)

		<-doneCh
	})
}
