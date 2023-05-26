package util

// A simple (non-synchronous) set using a map.
type Set[K comparable] struct {
	s map[K]struct{}
}

func (s *Set[K]) Add(key K) {
	s.s[key] = struct{}{}
}

func (s *Set[K]) Includes(key K) bool {
	_, ok := s.s[key]
	return ok
}
