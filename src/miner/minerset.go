package miner

import (
	"sync/atomic"
	"time"

	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

type MinerSet struct {
	exists       bool
	MinersActive atomic.Bool
	SolutionCh   <-chan kern.Block
	miners       []*Miner
}

func StartMinerSet(numMiners int) *MinerSet {
	if numMiners == 0 {
		return &MinerSet{}
	}
	chs := make([]chan kern.Block, numMiners)
	miners := make([]*Miner, numMiners)
	for i := 0; i < numMiners; i++ {
		miners[i] = NewMiner()
		go miners[i].Loop()
		chs[i] = miners[i].SolutionCh
	}
	aggSolutionCh := util.Aggregate(chs)
	minerSet := &MinerSet{
		exists:     true,
		SolutionCh: aggSolutionCh,
		miners:     miners,
	}
	minerSet.MinersActive.Store(true)
	return minerSet
}

func (ms *MinerSet) SetTargets(target kern.Block) {
	if !ms.exists {
		return
	}
	// Wait until miners ready
	delaySecs := 10
	ready := false
	for i := 0; i < delaySecs; i++ {
		ready = ms.MinersActive.Load()
		if ready {
			break
		}
		if i != delaySecs-1 {
			time.Sleep(time.Second)
		}
	}
	if !ready {
		return
	}
	// Set each target
	for i := 0; i < len(ms.miners); i++ {
		noisedTarget := target
		noise := kern.NewHashTRand()
		noisedTarget.Noise = noise
		ms.miners[i].SetTarget(noisedTarget)
	}
}
