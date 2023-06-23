package syncmap

import (
	"fmt"
	"sync"
)

// Generic allowing better type checking of sync.Map.
// This map intentionally does not support deletions.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

// Create a new SyncMap.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

// Check whether the given key exists in the map.
func (sm *SyncMap[K, V]) Has(key K) bool {
	_, ok := sm.m.Load(key)
	return ok
}

// Retrieve the given key from the map, panics if it doesn't exist.
func (sm *SyncMap[K, V]) Get(key K) V {
	v, ok := sm.m.Load(key)
	if !ok {
		panic(fmt.Sprint("expected key to exist:", key))
	}
	return v.(V)
}

// Store the given key in the map.
func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}
