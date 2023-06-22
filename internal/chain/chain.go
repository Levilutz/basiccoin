package chain

import (
	"fmt"

	"github.com/levilutz/basiccoin/internal/inv"
	"github.com/levilutz/basiccoin/internal/pubsub"
	"github.com/levilutz/basiccoin/pkg/topic"
	"github.com/levilutz/basiccoin/pkg/util"
)

// The chain's subscriptions.
// Ensure each of these is initialized in NewChain.
type subscriptions struct {
	// Events
	CandidateHead *topic.SubCh[pubsub.CandidateHeadEvent]
	CandidateTx   *topic.SubCh[pubsub.CandidateTxEvent]
	PrintUpdate   *topic.SubCh[pubsub.PrintUpdateEvent]
	// Queries
	PkhBalance *topic.SubCh[pubsub.PkhBalanceQuery]
	PkhUtxos   *topic.SubCh[pubsub.PkhUtxosQuery]
}

// A routine to manage our blockchain state and updates to it.
type Chain struct {
	pubSub        *pubsub.PubSub
	inv           *inv.Inv
	subs          *subscriptions
	state         *State
	supportMiners bool
}

// Create a new chain.
func NewChain(pubSub *pubsub.PubSub, inv *inv.Inv, supportMiners bool) *Chain {
	subs := &subscriptions{
		CandidateHead: pubSub.CandidateHead.SubCh(),
		CandidateTx:   pubSub.CandidateTx.SubCh(),
		PrintUpdate:   pubSub.PrintUpdate.SubCh(),
		PkhBalance:    pubSub.PkhBalance.SubCh(),
		PkhUtxos:      pubSub.PkhUtxos.SubCh(),
	}
	return &Chain{
		pubSub:        pubSub,
		inv:           inv,
		subs:          subs,
		state:         NewState(inv),
		supportMiners: supportMiners,
	}
}

// Start the chain's loop.
func (c *Chain) Loop() {
	if c.supportMiners {
		c.CreateMiningTarget()
	}
	for {
		select {
		case event := <-c.subs.CandidateHead.C:
			if err := c.handleCandidateHead(event); err != nil {
				fmt.Printf("failed to verify new chain: %s\n", err.Error())
			}

		case event := <-c.subs.CandidateTx.C:
			if err := c.handleCandidateTx(event); err != nil {
				fmt.Printf("failed to verify new tx: %s\n", err.Error())
			}

		case <-c.subs.PrintUpdate.C:
			fmt.Printf("chain height: %d\n", c.inv.GetBlockHeight(c.state.head))

		case query := <-c.subs.PkhBalance.C:
			util.WriteChIfPossible(query.Ret, c.state.GetPkhBalance(query.PublicKeyHash))

		case query := <-c.subs.PkhUtxos.C:
			util.WriteChIfPossible(query.Ret, c.state.GetPkhUtxos(query.PublicKeyHash))
		}
	}
}

// Upgrades our chain to the given new head, if it proves to be correct and better.
func (c *Chain) handleCandidateHead(event pubsub.CandidateHeadEvent) error {
	curHead := c.state.head
	// Insert each entity into the inventory, in order.
	for _, tx := range event.Txs {
		txId := tx.Hash()
		if !c.inv.HasTx(txId) {
			err := c.inv.StoreTx(tx)
			if err != nil {
				return err
			}
			c.state.AddMempoolTx(txId)
			// Don't re-broadcast tx directly, it's implicitly rebroadcasted with block
		}
	}
	for _, merkle := range event.Merkles {
		if !c.inv.HasMerkle(merkle.Hash()) {
			err := c.inv.StoreMerkle(merkle)
			if err != nil {
				return err
			}
		}
	}
	for _, block := range event.Blocks {
		if !c.inv.HasBlock(block.Hash()) {
			err := c.inv.StoreBlock(block)
			if err != nil {
				return err
			}
		}
	}
	// Verify new total work is higher
	if !c.inv.HasBlock(event.Head) {
		return fmt.Errorf("provided head not known and not provided")
	}
	newWork := c.inv.GetBlockTotalWork(event.Head)
	curWork := c.inv.GetBlockTotalWork(curHead)
	if !curWork.Lt(newWork) {
		return fmt.Errorf("new chain is not higher total work than current chain")
	}
	// Find common ancestor of our chain heads
	lcaId := c.inv.GetBlockLCA(curHead, event.Head)
	// Copy state, rewind to lca, and advance to head
	newState := c.state.Copy()
	newState.RewindUntil(lcaId)
	newBlocks := c.inv.GetBlockAncestorsUntil(event.Head, lcaId)
	// Advance through intermediate blocks, then the new head
	for i := len(newBlocks) - 1; i >= 0; i-- {
		if err := newState.Advance(newBlocks[i]); err != nil {
			return fmt.Errorf("failed to advance to block: %s", err.Error())
		}
	}
	if err := newState.Advance(event.Head); err != nil {
		return fmt.Errorf("failed to advance to block: %s", err.Error())
	}
	// Shift to new head - don't return error after here or state will get corrupted
	c.state = newState
	// Publish events
	c.pubSub.ValidatedHead.Pub(pubsub.ValidatedHeadEvent{
		Head: event.Head,
	})
	if c.supportMiners {
		c.CreateMiningTarget()
	}
	return nil
}

func (c *Chain) handleCandidateTx(event pubsub.CandidateTxEvent) error {
	txId := event.Tx.Hash()
	if c.inv.HasTx(txId) {
		return nil
	}
	if err := c.inv.StoreTx(event.Tx); err != nil {
		return err
	}
	c.state.AddMempoolTx(txId)
	c.pubSub.ValidatedTx.Pub(pubsub.ValidatedTxEvent{
		TxId: txId,
	})
	return nil
}
