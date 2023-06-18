package topic

import (
	"sync"
	"time"

	"github.com/levilutz/basiccoin/pkg/syncqueue"
)

// A subscription to read from (just a subset of SyncQueue methods).
type Sub[T any] interface {
	Pop() (T, bool)
	Close()
}

// A single pub-sub topic.
type Topic[T any] struct {
	mu     sync.RWMutex
	subs   map[uint64]*syncqueue.SyncQueue[T]
	nextId uint64
}

func NewTopic[T any]() *Topic[T] {
	topic := &Topic[T]{
		subs: make(map[uint64]*syncqueue.SyncQueue[T]),
	}
	// Start garbage collection loop
	go func() {
		gcTicker := time.NewTicker(time.Second * 15)
		for {
			<-gcTicker.C
			topic.gc()
		}
	}()
	return topic
}

// Garbage collect the topic's subscriptions.
func (t *Topic[T]) gc() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, sub := range t.subs {
		if sub.Closed() {
			delete(t.subs, id)
		}
	}
}

// Subscribe to this topic, return a subscription to read from.
func (t *Topic[T]) Sub() Sub[T] {
	t.mu.Lock()
	defer t.mu.Unlock()
	sub := syncqueue.NewSyncQueue[T]()
	t.subs[t.nextId] = sub
	t.nextId++
	return sub
}

// Publish messages to a topic, in the order given.
func (t *Topic[T]) Pub(msgs ...T) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, sub := range t.subs {
		sub.Push(msgs...)
	}
}
