package queue

// A simple (non-synchronous) queue using a slice.
type Queue[K any] struct {
	q []K
}

// Create a new queue.
func NewQueue[K any](keys ...K) *Queue[K] {
	return &Queue[K]{
		q: keys,
	}
}

// Push items to the end of the queue, in the order given.
func (q *Queue[K]) Push(keys ...K) {
	q.q = append(q.q, keys...)
}

// Pop an item off the front of the queue, if available.
func (q *Queue[K]) Pop() (key K, ok bool) {
	if len(q.q) == 0 {
		return key, false
	}
	key = q.q[0]
	q.q = q.q[1:]
	return key, true
}

// Get how many items are currently in the queue.
func (q *Queue[K]) Size() int {
	return len(q.q)
}
