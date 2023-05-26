package util

import "sync"

// Generic allowing better type checking of sync.Map.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func (sm *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := sm.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}
