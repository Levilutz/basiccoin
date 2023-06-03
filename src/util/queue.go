package util

type Queue[K any] struct {
	q []K
}

func NewQueue[K any](keys ...K) *Queue[K] {
	return &Queue[K]{
		q: keys,
	}
}

func (q *Queue[K]) Push(keys ...K) {
	q.q = append(q.q, keys...)
}

func (q *Queue[K]) Pop() (key K, ok bool) {
	if len(q.q) == 0 {
		return key, false
	}
	key = q.q[0]
	q.q = q.q[1:]
	return key, true
}

func (q *Queue[K]) Size() int {
	return len(q.q)
}
