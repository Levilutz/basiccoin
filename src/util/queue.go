package util

type Queue[K any] struct {
	q []K
}

func NewQueue[K any]() *Queue[K] {
	return &Queue[K]{
		q: make([]K, 0),
	}
}

func (q *Queue[K]) Push(key K) {
	q.q = append(q.q, key)
}

func (q *Queue[K]) Pop() (key K, ok bool) {
	if len(q.q) == 0 {
		return key, false
	}
	key = q.q[0]
	q.q = q.q[1:]
	return key, true
}
