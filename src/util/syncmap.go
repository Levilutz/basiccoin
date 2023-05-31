package util

import (
	"fmt"
	"sync"
)

// Generic allowing better type checking of sync.Map.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func (sm *SyncMap[K, V]) Has(key K) bool {
	_, ok := sm.m.Load(key)
	return ok
}

func (sm *SyncMap[K, V]) Get(key K) V {
	v, ok := sm.m.Load(key)
	if !ok {
		panic(fmt.Sprint("expected key to exist:", key))
	}
	return v.(V)
}

func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}
