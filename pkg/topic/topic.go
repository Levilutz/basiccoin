package topic

import (
	"sync"
	"sync/atomic"
)

// A single subscription to a pub-sub topic.
// The subscriber is obligated to read from the channel often enough to not block writers.
// Should only be read from / closed by a single thread.
// Will only be written to by a single Topic.Pub execution at once.
type Sub[T any] struct {
	C       chan T // A channel for the subscriber to read from.
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

// Post a message to this subscription. Blocks until successful. Returns whether to close.
func (s *Sub[T]) post(msg T) bool {
	if s.stopAck.Load() {
		// Should be unreachable so long as Topic doesn't post concurrently
		panic("attempt to post to closed subscription")
	}
	if s.stopReq.Load() {
		close(s.C)
		s.stopAck.Store(true)
		return true
	}
	s.C <- msg
	return false
}

// A single pub-sub topic.
type Topic[T any] struct {
	mu     sync.Mutex
	subs   map[uint64]*Sub[T]
	nextId uint64
}

func NewTopic[T any]() *Topic[T] {
	return &Topic[T]{
		subs: make(map[uint64]*Sub[T]),
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

// Publish a message to a topic.
func (t *Topic[T]) Pub(msg T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, sub := range t.subs {
		if sub.post(msg) {
			delete(t.subs, id)
		}
	}
}
