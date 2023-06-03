package util

// A simple (non-synchronous) set using a map.
type Set[K comparable] struct {
	s map[K]struct{}
}

// Create a new set.
func NewSet[K comparable]() *Set[K] {
	return &Set[K]{
		s: make(map[K]struct{}),
	}
}

// Create a new set from a list.
func NewSetFromList[K comparable](list []K) *Set[K] {
	s := NewSet[K]()
	for _, item := range list {
		s.Add(item)
	}
	return s
}

// Add a key to the set.
func (s *Set[K]) Add(keys ...K) {
	for _, key := range keys {
		s.s[key] = struct{}{}
	}
}

// Remove a key from the set, return whether it existed.
func (s *Set[K]) Remove(key K) bool {
	if !s.Includes(key) {
		return false
	}
	delete(s.s, key)
	return true
}

// Check whether a set includes the given item.
func (s *Set[K]) Includes(key K) bool {
	_, ok := s.s[key]
	return ok
}

// Check whether a set includes any of the given items.
func (s *Set[K]) IncludesAny(keys ...K) bool {
	for _, key := range keys {
		if s.Includes(key) {
			return true
		}
	}
	return false
}

// Shallow copy the set.
func (s *Set[K]) Copy() *Set[K] {
	newS := make(map[K]struct{}, len(s.s))
	for k := range s.s {
		newS[k] = struct{}{}
	}
	return &Set[K]{s: newS}
}

// Filter based on a given func. If filter returns false, element is removed.
func (s *Set[K]) Filter(filter func(key K) bool) {
	for key := range s.s {
		if !filter(key) {
			delete(s.s, key)
		}
	}
}

// Create a list from the set.
func (s *Set[K]) ToList() []K {
	out := make([]K, len(s.s))
	i := 0
	for k := range s.s {
		out[i] = k
		i++
	}
	return out
}

// Check whether this set intersects another.
func (s *Set[K]) HasIntersection(other *Set[K]) bool {
	for item := range other.s {
		if s.Includes(item) {
			return true
		}
	}
	return false
}
