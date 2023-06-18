package topic

import (
	"sync"
	"sync/atomic"
)

// A single subscription to a pub-sub topic.
// The subscriber is obligated to read from the channel before buffer fills.
// Should only be read from / closed by a single thread.
type Sub[T any] struct {
	C       chan T     // A channel for the subscriber to read from.
	mu      sync.Mutex // To prevent simultaneous posts
	stopReq atomic.Bool
	stopAck atomic.Bool
}

// Close the subscription. Blocks until successful.
// Cannot run simultaneously with other Closes or reads on C.
func (s *Sub[T]) Close() {
	s.stopReq.Store(true)
	for ok := true; !s.stopAck.Load() || ok; {
		_, ok = <-s.C
	}
}

// Post a message to this subscription. Blocks until successful.
func (s *Sub[T]) post(msg T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopAck.Load() {
		return
	} else if s.stopReq.Load() {
		close(s.C)
		s.stopAck.Store(true)
	} else {
		s.C <- msg
	}
}

// A single pub-sub topic.
type Topic[T any] struct {
	mu     sync.RWMutex
	subs   map[uint64]*Sub[T]
	nextId uint64
}

func NewTopic[T any]() *Topic[T] {
	topic := &Topic[T]{
		subs: make(map[uint64]*Sub[T]),
	}
	// Start garbage collection loop
	go func() {
		for {
			topic.gc()
		}
	}()
	return topic
}

// Garbage collect the topic's subscriptions.
func (t *Topic[T]) gc() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, sub := range t.subs {
		if sub.stopAck.Load() {
			delete(t.subs, id)
		}
	}
}

// Subscribe to this topic, return a subscription to read from.
// The subscriber is obligated to read from the channel often enough to not block pubs.
func (t *Topic[T]) Sub() *Sub[T] {
	t.mu.Lock()
	defer t.mu.Unlock()
	sub := &Sub[T]{
		C: make(chan T, 1),
	}
	t.subs[t.nextId] = sub
	t.nextId++
	return sub
}

// Publish a message to a topic. Blocks until successfully written to each channel.
func (t *Topic[T]) Pub(msg T) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, sub := range t.subs {
		sub.post(msg)
	}
}
