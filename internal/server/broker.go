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
			b.withLock(func() {
				b.clients[c] = true
			})

		case c := <-b.unregister:
			b.withLock(func() {
				if _, ok := b.clients[c]; ok {
					delete(b.clients, c)
					close(c.send)
				}
			})

		case msg := <-b.broadcast:
			b.withRLock(func() {
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
			})
		}
	}
}

func (b *wsBroker) withLock(f func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	f()
}

func (b *wsBroker) withRLock(f func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	f()
}
