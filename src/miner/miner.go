package miner

import (
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/db"
)

type Miner struct {
	target      db.Block
	newTargetCh chan db.Block
	killCh      chan struct{}
	solutionCh  chan<- db.Block
	nextNonce   uint32
}

func NewMiner(target db.Block, solutionCh chan<- db.Block) *Miner {
	return &Miner{
		target:      target,
		newTargetCh: make(chan db.Block),
		killCh:      make(chan struct{}),
		solutionCh:  solutionCh,
		nextNonce:   0,
	}
}

// Set a new target to be handled.
func (m *Miner) SetTarget(block db.Block) {
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
	defer fmt.Println("Miner closed")
	for {
		select {
		case newTarget := <-m.newTargetCh:
			m.target = newTarget
			m.nextNonce = 0

		case <-m.killCh:
			return

		default:
			if m.nextNonce == 1<<32-1 {
				time.Sleep(time.Second)
			} else {
				solution, ok := m.mine(1 << 16)
				if ok {
					m.solutionCh <- solution
				}
			}
		}
	}
}

// Keep trying nonces until it hits 2^32-1, then quit.
func (m *Miner) mine(rounds uint32) (db.Block, bool) {
	for i := uint32(0); i < rounds; i++ {
		target := db.Block{
			PrevBlockId: m.target.PrevBlockId,
			MerkleRoot:  m.target.MerkleRoot,
			Difficulty:  m.target.Difficulty,
			Nonce:       m.nextNonce,
		}
		hash := target.Hash()
		if m.nextNonce != 1<<32-1 {
			m.nextNonce += 1
		}
		if db.BelowTarget(hash, m.target.Difficulty) {
			return target, true
		}
		if m.nextNonce == 1<<32-1 {
			return db.Block{}, false
		}
	}
	return db.Block{}, false
}
