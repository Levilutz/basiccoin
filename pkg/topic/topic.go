package topic

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/levilutz/basiccoin/pkg/syncqueue"
)

// A subscription to read from (just a subset of SyncQueue methods).
type Sub[T any] interface {
	Pop() (T, bool)
	Peek() (T, bool)
	Size() int
	Close()
}

// A single pub-sub topic.
type Topic[T any] struct {
	mu     sync.RWMutex
	subs   map[uint64]*syncqueue.SyncQueue[T]
	nextId uint64
}

// Create a new topic with its gc routine already started.
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

// Subscribe to the topic, return a subscription channel.
func (t *Topic[T]) SubCh() *SubCh[T] {
	subCh := &SubCh[T]{
		Sub:      make(chan T),
		subQueue: t.Sub(),
		close:    make(chan struct{}),
	}
	go subCh.loop()
	return subCh
}

// A subscription channel.
type SubCh[T any] struct {
	Sub      chan T
	subQueue Sub[T]
	close    chan struct{}
}

// Close this subscription channel.
func (s *SubCh[T]) Close() {
	// Pull items off the channel until our close is processed.
	done := atomic.Bool{}
	go func() {
		s.close <- struct{}{}
		done.Store(true)
	}()
	for !done.Load() {
		<-s.Sub
	}
	close(s.Sub)
}

// Loop taking items from the queue and pushing them to the channel.
func (s *SubCh[T]) loop() {
	for {
		select {
		case <-s.close:
			s.subQueue.Close()
			return
		default:
			var msg T
			ok := true
			for ok {
				msg, ok = s.subQueue.Pop()
				s.Sub <- msg
			}
			time.Sleep(time.Millisecond * 25)
		}
	}
}
