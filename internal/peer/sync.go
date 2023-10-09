package peer

import (
	"fmt"

	"github.com/levilutz/basiccoin/internal/bus"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/queue"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/util"
)

var syncChainCmd = "sync-chain"

// Handle a chain sync, inbound or outbound.
func (p *Peer) handleSyncChain() error {
	ourWork := p.inv.GetBlockTotalWork(p.curHead)
	p.conn.WriteHashT(ourWork)
	p.conn.WriteHashT(p.curHead)
	theirWork := p.conn.ReadHashT()
	theirHead := p.conn.ReadHashT()
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Even if our work mismatches, we might have their head
	// This could mean manager is currently including that head, or it failed to before.
	if theirWork.Eq(ourWork) || (ourWork.Lt(theirWork) && p.inv.HasBlock(theirHead)) {
		p.conn.WriteString("cancel")
		p.conn.ReadString() // Just to consume their msg
		return p.conn.Err()
	} else {
		p.conn.WriteString("continue")
		resp := p.conn.ReadString()
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		if resp == "cancel" {
			return nil
		} else if resp != "continue" {
			return fmt.Errorf("unexpected peer response: %s", resp)
		}
	}
	// Someone wants a sync
	if theirWork.Lt(ourWork) {
		// Send a sync
		p.conn.WriteString("sync:send")
		p.conn.ReadStringExpected("sync:recv")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return p.handleOutboundSyncChain()
	} else {
		// Receive a sync
		p.conn.WriteString("sync:recv")
		p.conn.ReadStringExpected("sync:send")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return p.handleInboundSyncChain()
	}
}

// Handle an outbound chain sync.
func (p *Peer) handleOutboundSyncChain() error {
	// Find last common ancestor
	neededBlockIds := make([]core.HashT, 0)
	lcaId := p.curHead
	p.conn.WriteHashT(lcaId)
	resp := p.conn.ReadBool()
	for !p.conn.HasErr() && !resp {
		neededBlockIds = append(neededBlockIds, lcaId)
		lcaId = p.inv.GetBlockParentId(lcaId)
		p.conn.WriteHashT(lcaId)
		resp = p.conn.ReadBool()
	}
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if len(neededBlockIds) == 0 {
		return fmt.Errorf("peer does not need upgrade, sync should not have run")
	}
	// Send blocks to peer
	for _, blockId := range neededBlockIds {
		p.conn.WriteBlock(p.inv.GetBlock(blockId))
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	// Check if peer verified our work
	resp = p.conn.ReadBool()
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if !resp {
		return fmt.Errorf("peer failed to verify our chain")
	}
	// Send peer the entities it doesn't know about
	for id := p.conn.ReadHashT(); !p.conn.HasErr() && !id.EqZero(); id = p.conn.ReadHashT() {
		if p.inv.HasMerkle(id) {
			p.conn.WriteBool(false)
			p.conn.WriteMerkle(p.inv.GetMerkle(id))
		} else if p.inv.HasTx(id) {
			p.conn.WriteBool(true)
			p.conn.WriteTx(p.inv.GetTx(id))
		} else {
			return fmt.Errorf("peer requested unknown entity %s", id)
		}
	}
	p.conn.ReadStringExpected("complete")
	return p.conn.Err() // This catches err in entity negotiation too
}

// Handle an inbound chain sync.
func (p *Peer) handleInboundSyncChain() error {
	// Find last common ancestor
	neededBlockIds := make([]core.HashT, 0)
	newHead := p.conn.ReadHashT()
	lcaId := newHead
	for !p.conn.HasErr() && !p.inv.HasBlock(lcaId) {
		neededBlockIds = append(neededBlockIds, lcaId)
		p.conn.WriteBool(false)
		lcaId = p.conn.ReadHashT()
	}
	p.conn.WriteBool(true)
	if p.conn.HasErr() {
		return p.conn.Err()
	} else if len(neededBlockIds) == 0 {
		return fmt.Errorf("we do not actually need upgrade, sync should not have run")
	}
	// Receive blocks from peer
	newBlocks := make([]core.Block, len(neededBlockIds))
	for i, blockId := range neededBlockIds {
		newBlocks[i] = p.conn.ReadBlock(blockId)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	// Verify proof of work
	if err := p.quickVerifyChain(newHead, lcaId, newBlocks); err != nil {
		p.conn.WriteBool(false)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return fmt.Errorf("failed to verify received chain: %s", p.conn.Err().Error())
	}
	p.conn.WriteBool(true)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Request the peer for entities we don't know about
	newMerkles := make([]core.MerkleNode, 0)
	newTxs := make([]core.Tx, 0)
	idQueue := queue.NewQueue[core.HashT]()
	receivedIds := set.NewSet[core.HashT]()
	for _, block := range newBlocks {
		idQueue.Push(block.MerkleRoot)
	}
	for idQueue.Size() > 0 {
		id, _ := idQueue.Pop()
		if p.inv.HasEntity(id) || receivedIds.Includes(id) {
			continue
		}
		p.conn.WriteHashT(id)
		isTx := p.conn.ReadBool()
		if p.conn.HasErr() {
			return p.conn.Err()
		} else if !isTx {
			merkle := p.conn.ReadMerkle(id)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			newMerkles = append(newMerkles, merkle)
			idQueue.Push(merkle.LChild, merkle.RChild)
		} else {
			tx := p.conn.ReadTx(id)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			newTxs = append(newTxs, tx)
		}
		receivedIds.Add(id)
	}
	p.conn.WriteHashT(core.HashT{})
	p.conn.WriteString("complete")
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Send the new entities to the head (reversed so they're inserted in correct order)
	p.bus.CandidateHead.Pub(bus.CandidateHeadEvent{
		Head:                   newHead,
		Blocks:                 util.Reverse(newBlocks),
		Merkles:                util.Reverse(newMerkles),
		Txs:                    util.Reverse(newTxs),
		AutoAddMempoolInsecure: false,
	})
	return nil
}

// Verify a new chain's continuity, expected endpoints, and proof-of-work.
// Leaves havier and state-based verification to chain, this just helps prevent dos attacks.
func (p *Peer) quickVerifyChain(
	newHead core.HashT, lcaId core.HashT, newBlocks []core.Block,
) error {
	// This could be sped up by caching hashes, but it's not significant
	// Verify chain has expected head
	if newBlocks[0].Hash() != newHead {
		return fmt.Errorf("does not have expected head")
	}
	// Verify chain continuous
	for i := 0; i < len(newBlocks)-1; i++ {
		if newBlocks[i].PrevBlockId != newBlocks[i+1].Hash() {
			return fmt.Errorf("not continuous")
		}
	}
	// Verify chain attaches to lca
	if newBlocks[len(newBlocks)-1].PrevBlockId != lcaId {
		return fmt.Errorf("does not attach to last common ancestor")
	}
	// Verify claimed total work beats ours
	newWork := p.inv.GetBlockTotalWork(lcaId)
	for _, block := range newBlocks {
		newWork = newWork.WorkAppendTarget(block.Target)
	}
	if !p.inv.GetBlockTotalWork(p.curHead).Lt(newWork) {
		return fmt.Errorf("does not have higher proven work than our current chain")
	}
	// Verify each block actually beats claimed target
	for _, block := range newBlocks {
		if !block.Hash().Lt(block.Target) {
			return fmt.Errorf("block does not beat claimed target")
		}
	}
	return nil
}
