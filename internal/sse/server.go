package sse

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Simple SSE broker
type Broker struct {
	clients map[chan string]bool
	lock    sync.Mutex
	// buffer of recent events (raw JSON strings)
	buffer []string // JSON-encoded events, in order
	out    chan string
}

func NewBroker() *Broker {
	b := &Broker{
		clients: make(map[chan string]bool),
		buffer:  make([]string, 0, 128),
		out:     make(chan string, 256),
	}
	go b.runBroadcaster()
	return b
}

// DefaultBroker is a shared broker instance used by the application
var DefaultBroker = NewBroker()

// PublishEvent is a convenience wrapper to publish to the default broker
func PublishEvent(v interface{}) {
	DefaultBroker.Publish(v)
}

// Subscribe registers a client and returns a channel to receive messages
func (b *Broker) Subscribe() chan string {
	b.lock.Lock()
	defer b.lock.Unlock()
	ch := make(chan string, 10)
	b.clients[ch] = true
	return ch
}

// Unsubscribe removes a client channel
func (b *Broker) Unsubscribe(ch chan string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.clients, ch)
	close(ch)
}

// Publish sends message to the broker (non-blocking). If out buffer is full, event is dropped.
func (b *Broker) Publish(v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Printf("sse: marshal error: %v", err)
		return
	}
	s := string(bs)

	select {
	case b.out <- s:
	default:
		log.Printf("sse: dropping event, out buffer full")
	}
}

// broadcaster runs in background: append to buffer and fan out to clients non-blocking
func (b *Broker) runBroadcaster() {
	for raw := range b.out {
		b.lock.Lock()
		b.buffer = append(b.buffer, raw)
		if len(b.buffer) > cap(b.buffer) {
			b.buffer = b.buffer[len(b.buffer)-cap(b.buffer):]
		}

		for ch := range b.clients {
			select {
			case ch <- raw:
			default:
				// drop for slow client
			}
		}
		b.lock.Unlock()
	}
}

// Handler returns an http.Handler that serves SSE events
func (b *Broker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := b.Subscribe()
		defer b.Unsubscribe(ch)

		// Send a comment to establish the stream
		_, _ = w.Write([]byte(":ok\n\n"))
		flusher.Flush()

		// Replay buffer (simple behavior: send entire buffer)

		notify := r.Context().Done()

		for {
			select {
			case <-notify:
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				_, _ = w.Write([]byte("data: " + msg + "\n\n"))
				flusher.Flush()
			}
		}
	}
}
