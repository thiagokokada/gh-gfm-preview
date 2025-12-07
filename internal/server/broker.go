package server

import "sync"

// wsBroker handles registering/unregistering clients and broadcasting messages to them.
type wsBroker struct {
	// Registered clients.
	clients map[*wsClient]bool
	// Register requests from the clients.
	register chan *wsClient
	// Unregister requests from clients.
	unregister chan *wsClient
	// Broadcast messages to all clients.
	broadcast chan []byte
	// Protect clients map during iteration if needed elsewhere.
	mu sync.RWMutex
}

func newBroker() *wsBroker {
	return &wsBroker{
		clients:    make(map[*wsClient]bool),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		broadcast:  make(chan []byte, 1),
	}
}

func (b *wsBroker) run() {
	for {
		select {
		case c := <-b.register:
			b.mu.Lock()
			b.clients[c] = true
			b.mu.Unlock()
		case c := <-b.unregister:
			b.mu.Lock()

			if _, ok := b.clients[c]; ok {
				delete(b.clients, c)
				close(c.send)
			}

			b.mu.Unlock()
		case msg := <-b.broadcast:
			b.mu.RLock()

			for c := range b.clients {
				// try sending without blocking the broker
				select {
				case c.send <- msg:
				default:
					// client send channel is full or blocked; unregister it to avoid leaking
					b.mu.RUnlock()

					b.unregister <- c

					b.mu.RLock()
				}
			}

			b.mu.RUnlock()
		}
	}
}
