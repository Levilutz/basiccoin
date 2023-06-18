package miner

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/kern"
)

type Miner struct {
	SolutionCh  chan kern.Block
	target      kern.Block
	newTargetCh chan kern.Block
	killCh      chan struct{}
	nextNonce   uint64
}

func NewMiner() *Miner {
	return &Miner{
		target:      kern.Block{},
		newTargetCh: make(chan kern.Block),
		killCh:      make(chan struct{}),
		SolutionCh:  make(chan kern.Block),
		nextNonce:   0,
	}
}

// Set a new target to be handled.
func (m *Miner) SetTarget(block kern.Block) {
	go func() {
		m.newTargetCh <- block
	}()
}

// Kill the miner
func (m *Miner) Kill() {
	go func() {
		m.killCh <- struct{}{}
	}()
}

// Loop handling events and searching for block solutions.
func (m *Miner) Loop() {
	defer fmt.Println("miner closed")
	for {
		select {
		case newTarget := <-m.newTargetCh:
			m.target = newTarget
			m.nextNonce = 0

		case <-m.killCh:
			return

		default:
			if m.target.MerkleRoot.EqZero() || m.nextNonce == 1<<64-1 {
				time.Sleep(time.Second)
			} else {
				solution, ok := m.mine(1 << 16)
				if ok {
					m.SolutionCh <- solution
				}
			}
		}
	}
}

// Keep trying nonces until it hits 2^64-1, then quit.
func (m *Miner) mine(rounds uint64) (kern.Block, bool) {
	for i := uint64(0); i < rounds; i++ {
		target := kern.Block{
			PrevBlockId: m.target.PrevBlockId,
			MerkleRoot:  m.target.MerkleRoot,
			Target:      m.target.Target,
			Noise:       m.target.Noise,
			Nonce:       m.nextNonce,
			MinedTime:   m.target.MinedTime,
		}
		hash := target.Hash()
		if m.nextNonce != 1<<64-1 {
			m.nextNonce += 1
		}
		if hash.Lt(m.target.Target) {
			return target, true
		}
		if m.nextNonce == 1<<64-1 {
			return kern.Block{}, false
		}
	}
	return kern.Block{}, false
}
