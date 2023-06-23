package topic_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/levilutz/basiccoin/pkg/topic"
	"github.com/levilutz/basiccoin/pkg/util"
)

// Test topic with multiple publishers and multiple subscribers.
func TestTopic(t *testing.T) {
	var wg sync.WaitGroup
	topic := NewTopic[string]()
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 5)
		wg.Add(1)
		sub := topic.Sub()
		i := i
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 10)
			for j := 0; j < i; j++ {
				msg, ok := sub.Pop()
				t.Logf("sub %d received '%s' (%t)", i, msg, ok)
			}
			sub.Close()
			topic.Pub(fmt.Sprintf("sub %d closed", i))
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			topic.Pub(fmt.Sprint("message ", i))
		}()
	}
	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	timer := time.NewTimer(time.Second)
	select {
	case <-done:
		return
	case <-timer.C:
		util.Assert(t, false, "time out error")
	}
}
