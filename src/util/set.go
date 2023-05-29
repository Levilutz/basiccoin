package util

// A simple (non-synchronous) set using a map.
type Set[K comparable] struct {
	s map[K]struct{}
}

func NewSet[K comparable]() *Set[K] {
	return &Set[K]{
		s: make(map[K]struct{}),
	}
}

func (s *Set[K]) Add(key K) {
	s.s[key] = struct{}{}
}

func (s *Set[K]) Includes(key K) bool {
	_, ok := s.s[key]
	return ok
}

func (s *Set[K]) Copy() *Set[K] {
	newS := make(map[K]struct{}, len(s.s))
	for k := range s.s {
		newS[k] = struct{}{}
	}
	return &Set[K]{s: newS}
}
