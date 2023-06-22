package miner

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/topic"
	"github.com/levilutz/basiccoin/pkg/util"
)

// The miner's subscriptions.
// Ensure each of these is initialized in NewMiner.
type subscriptions struct {
	MinerTarget *topic.SubCh[bus.MinerTargetEvent]
}

// Close our subscriptions as we close.
func (s subscriptions) Close() {
	s.MinerTarget.Close()
}

// A single-threaded miner instance.
type Miner struct {
	params      Params
	bus         *bus.Bus
	inv         inv.InvReader
	subs        *subscriptions
	template    *core.Block
	outCoinbase *core.Tx
	outMerkles  []core.MerkleNode
}

// Create a new miner.
func NewMiner(params Params, msgBus *bus.Bus, inv inv.InvReader) *Miner {
	subs := &subscriptions{
		MinerTarget: msgBus.MinerTarget.SubCh(),
	}
	return &Miner{
		params:      params,
		bus:         msgBus,
		inv:         inv,
		subs:        subs,
		template:    nil,
		outCoinbase: nil,
		outMerkles:  nil,
	}
}

// Start the miner's loop.
func (m *Miner) Loop() {
	// Handle panics and unsubscribe
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("miner closed from panic: %s\n", r)
			debug.PrintStack()
		} else {
			fmt.Println("miner closed")
		}
		m.subs.Close()
	}()

	// Loop
	for {
		select {
		case event := <-m.subs.MinerTarget.C:
			m.updateTemplate(event.Head, event.Target, event.TxIds)

		default:
			if m.template == nil {
				time.Sleep(time.Millisecond * 100)
			} else {
				m.mine(1 << 20)
			}
		}
	}
}

// Update our stored template and outputs from the given target data.
func (m *Miner) updateTemplate(head core.HashT, target core.HashT, txIds []core.HashT) {
	// Compute total fees
	totalFees := uint64(0)
	for _, txId := range txIds {
		tx := m.inv.GetTx(txId)
		if !tx.HasSurplus() {
			panic("miner was given tx with negative surplus")
		}
		totalFees += tx.InputsValue() - tx.OutputsValue()
	}
	// Make coinbase tx
	coinbaseTx := core.Tx{
		IsCoinbase: true,
		MinBlock:   m.inv.GetBlockHeight(head) + 1,
		Inputs:     make([]core.TxIn, 0),
		Outputs: []core.TxOut{
			{
				Value:         totalFees + m.inv.GetCoreParams().BlockReward,
				PublicKeyHash: m.params.PayoutPkh,
			},
		},
	}
	// Build merkle tree from txs
	txIds = util.Prepend(txIds, coinbaseTx.Hash())
	merkleMap, merkleIds := core.MerkleFromTxIds(txIds)
	outMerkles := make([]core.MerkleNode, len(merkleIds))
	for i, merkleId := range merkleIds {
		outMerkles[i] = merkleMap[merkleId]
	}
	// Set our template and output data
	m.template = &core.Block{
		PrevBlockId: head,
		MerkleRoot:  merkleIds[len(merkleIds)-1],
		Target:      target,
		Noise:       core.NewHashTRand(),
		Nonce:       0,
		MinedTime:   uint64(time.Now().Unix()),
	}
	m.outCoinbase = &coinbaseTx
	m.outMerkles = outMerkles
}

// Try nonces for the specified number of rounds.
func (m *Miner) mine(rounds uint64) {
	for i := uint64(0); i < rounds; i++ {
		hash := m.template.Hash()
		if hash.Lt(m.template.Target) {
			m.publishSolution()
		}
		if m.template.Nonce == 1<<64-1 {
			m.template.Noise = core.NewHashTRand()
			m.template.Nonce = 0
		} else {
			m.template.Nonce++
		}
	}
}

// Publish the currently held solution.
func (m *Miner) publishSolution() {
	fmt.Println("!!! mined solution")
	m.bus.CandidateHead.Pub(bus.CandidateHeadEvent{
		Head:    m.template.Hash(),
		Blocks:  []core.Block{*m.template},
		Merkles: util.CopyList(m.outMerkles),
		Txs:     []core.Tx{*m.outCoinbase},
	})
}
