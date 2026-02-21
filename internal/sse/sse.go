// Package sse provides Server-Sent Events support.
package sse

import "sync"

// Broker manages SSE client channels for broadcasting reload events.
type Broker interface {
	// Register creates and registers a new client channel, returning it for reading.
	Register() <-chan struct{}
	// Unregister removes the client channel from the broker and closes it.
	Unregister(ch <-chan struct{})
	// Broadcast sends a reload signal to all registered clients (non-blocking).
	Broadcast()
}

type broker struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

// New creates a new Broker.
func New() Broker {
	return &broker{
		clients: make(map[chan struct{}]struct{}),
	}
}

func (b *broker) Register() <-chan struct{} {
	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *broker) Unregister(receive <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		if (<-chan struct{})(ch) == receive {
			delete(b.clients, ch)
			close(ch)
			return
		}
	}
}

func (b *broker) Broadcast() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- struct{}{}:
		default:
			// クライアントのバッファが満杯の場合はスキップ（ノンブロッキング）
		}
	}
}
