package syncqueue

import (
	"sync"

	"github.com/levilutz/basiccoin/pkg/queue"
)

// A simple sync queue.
type SyncQueue[K any] struct {
	mu     sync.Mutex
	q      *queue.Queue[K]
	closed bool
}

// Create a new SyncQueue.
func NewSyncQueue[K any](keys ...K) *SyncQueue[K] {
	return &SyncQueue[K]{
		q: queue.NewQueue[K](keys...),
	}
}

// Push items to the end of the queue, in the order given.
func (q *SyncQueue[K]) Push(keys ...K) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.q.Push(keys...)
}

// Pop an item off the front of the queue, if available.
func (q *SyncQueue[K]) Pop() (key K, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return key, false
	}
	return q.q.Pop()
}

// Get how many items are currently in the queue.
func (q *SyncQueue[K]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return 0
	}
	return q.q.Size()
}

// Close the queue to further reads and writes.
func (q *SyncQueue[K]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.q = nil
}

// Check whether the queue is closed.
func (q *SyncQueue[K]) Closed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closed
}
