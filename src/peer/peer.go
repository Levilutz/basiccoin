package peer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/kern"
	"github.com/levilutz/basiccoin/src/util"
)

type MainEventHandler interface {
	HandlePeerClosing(runtimeId string)
	HandleInboundSync(
		head kern.HashT,
		blocks []kern.Block,
		merkles []kern.MerkleNode,
		txs []kern.Tx,
	)
	HandlePeersReceived(addrs []string)
	HandlePeersWanted(runtimeId string)
	HandleNewTx(tx kern.Tx)
}

// Encapsulate a high-level connection to a peer.
type Peer struct {
	Info           *PeerInfo
	eventBus       chan any
	conn           *PeerConn
	mainHandler    MainEventHandler
	weAreInitiator bool
	inv            db.InvReader
	head           kern.HashT
	shouldClose    bool
}

// Create a Peer.
// "msg" is the result of a successful handshake on a PeerConn.
// "pc" is the peerconn on which we have already handshaked.
// "mainBus" is a bus on which to emit events back to the manager.
// "weAreInitiator" is whether we are the peer that initiated the connection.
// "inv" is a InvReader.
func NewPeer(
	info *PeerInfo,
	pc *PeerConn,
	mainHandler MainEventHandler,
	weAreInitiator bool,
	inv db.InvReader,
	head kern.HashT,
) *Peer {
	return &Peer{
		Info:           info,
		eventBus:       make(chan any),
		conn:           pc,
		mainHandler:    mainHandler,
		weAreInitiator: weAreInitiator,
		inv:            inv,
		head:           head,
	}
}

func (p *Peer) queueEvent(event any) {
	go func() { p.eventBus <- event }()
}

// Tell the peer it should terminate the connection.
func (p *Peer) ShouldEnd() {
	p.queueEvent(shouldEndEvent{})
}

// Inform this peer of a new kern.HashTad.
func (p *Peer) SyncHead(head kern.HashT) {
	p.queueEvent(syncHeadEvent{
		head: head,
	})
}

// Inform the peer of other peers.
func (p *Peer) SendPeersData(addrs []string) {
	p.queueEvent(peersDataEvent{
		addrs: addrs,
	})
}

// Request that the peer send back other addrs.
func (p *Peer) PeersWanted() {
	p.queueEvent(peersWantedEvent{})
}

func (p *Peer) SendTx(txId kern.HashT) {
	p.queueEvent(sendTxEvent{
		txId: txId,
	})
}

// Loop handling events from our message bus and the peer.
func (p *Peer) Loop() {
	defer func() {
		fmt.Println("peer closed:", p.Info.RuntimeID)
		p.mainHandler.HandlePeerClosing(p.Info.RuntimeID)
	}()
	pingTicker := time.NewTicker(util.Constants.PeerPingFreq)
	for {
		select {
		case event := <-p.eventBus:
			if err := p.handlePeerBusEvent(event); err != nil {
				fmt.Printf("error handling event '%T': %s\n", event, err.Error())
			}

		case <-pingTicker.C:
			if err := p.issuePeerCommand("ping", func() error { return nil }); err != nil {
				fmt.Println("peer lost:", p.Info.RuntimeID, err.Error())
				return
			}

		default:
			line := p.conn.ReadLineTimeout(100 * time.Millisecond)
			if p.conn.HasErr() {
				p.conn.Err() // Drop it
				continue
			}
			if err := p.handleReceivedLine(line); err != nil {
				fmt.Printf("error handling line '%s': %s\n", line, err.Error())
			}
		}
		if p.shouldClose {
			return
		}
	}
}

// Handle event from our message bus, return whether we should close.
func (p *Peer) handlePeerBusEvent(event any) error {
	switch msg := event.(type) {
	case shouldEndEvent:
		p.shouldClose = true
		return p.handleClose(true)

	case syncHeadEvent:
		p.head = msg.head
		return p.issuePeerCommand("sync", p.handleSync)

	case peersDataEvent:
		return p.issuePeerCommand("addrs", func() error {
			p.conn.TransmitUint64Line(uint64(len(msg.addrs)))
			for _, addr := range msg.addrs {
				p.conn.TransmitStringLine(addr)
			}
			p.conn.TransmitStringLine("fin:addrs")
			return p.conn.Err()
		})

	case peersWantedEvent:
		return p.issuePeerCommand("peers-wanted", func() error {
			return nil
		})

	case sendTxEvent:
		return p.issuePeerCommand("tx", func() error {
			p.conn.TransmitHashLine(msg.txId)
			resp := p.conn.RetryReadStringLine(7)
			if resp == "next" {
				p.conn.TransmitTx(p.inv.GetTx(msg.txId))
			}
			return p.conn.Err()
		})

	default:
		fmt.Printf("unhandled peer event %T\n", event)
	}
	return nil
}

// Handle command received from peer, returns whether we should close.
func (p *Peer) handleReceivedLine(line []byte) error {
	if !bytes.HasPrefix(line, []byte("cmd:")) {
		return fmt.Errorf("unrecognized line: %s", line)
	}
	command := string(line)[4:]
	if command == "close" {
		p.shouldClose = true
		return p.handleClose(false)
	}

	p.conn.TransmitStringLine("ack:" + command)
	if p.conn.HasErr() {
		return p.conn.Err()
	}

	// TODO: Abstract all these into a dispatch? or attach handlers http library style?
	if command == "ping" {
		return nil

	} else if command == "sync" {
		return p.handleSync()

	} else if command == "addrs" {
		return p.handleReceiveAddrs()

	} else if command == "peers-wanted" {
		return p.handleReceivePeersWanted()

	} else if command == "tx" {
		return p.handleReceiveTx()

	} else {
		return fmt.Errorf("unexpected peer message: %s", command)
	}
}

// Issue an outbound interaction for the command (given without "cmd:").
// Handler is what to run after they ack. Returns whether we should close.
// If us and peer simultaneously issued commands, the og handshake initiator goes last.
func (p *Peer) issuePeerCommand(command string, handler func() error) error {
	p.conn.TransmitStringLine("cmd:" + command)
	// Expect to receive either "ack:our command" or "cmd:their command"
	resp := p.conn.RetryReadLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Happy path - they acknowledged us
	if string(resp) == "ack:"+command {
		return handler()
	}
	// Sad path - we sent commands simultaneously
	if bytes.HasPrefix(resp, []byte("cmd:")) {
		if string(resp) == "cmd:close" {
			// If their command was a close, handle it immediately
			p.shouldClose = true
			return p.handleClose(false)

		} else if p.weAreInitiator {
			// If we initiated the og handshake, honor their cmd, then expect ours to be
			if err := p.handleReceivedLine(resp); p.shouldClose || err != nil {
				return err
			}
			p.conn.ConsumeExpected("ack:" + command)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			return handler()

		} else {
			// If we received the og handshake, expect to be honored, then honor theirs
			p.conn.ConsumeExpected("ack:" + command)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			err := handler()
			if err != nil {
				return err
			}
			return p.handleReceivedLine(resp)
		}
	}
	return nil
}

func (p *Peer) handleClose(issuing bool) error {
	if issuing {
		p.conn.TransmitStringLine("cmd:close")
	}
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	return p.conn.Close()
}

// Handle a sync, inbound or outbound.
// If we're successful and we received the sync, send InboundSyncMainEvent to main.
// If we're successful and we sent the sync or no sync occurred, send nothing to main.
func (p *Peer) handleSync() error {
	ourWork := p.inv.GetBlockTotalWork(p.head)
	p.conn.TransmitHashLine(ourWork)
	p.conn.TransmitHashLine(p.head)
	theirWork := p.conn.RetryReadHashLine(7)
	theirHead := p.conn.RetryReadHashLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Even if our work mismatches, we might have their head
	// This could mean manager is currently including that head, or it failed to before.
	if theirWork.Eq(ourWork) ||
		(ourWork.Lt(theirWork) && p.inv.HasBlock(theirHead)) {
		p.conn.TransmitStringLine("fin:sync")
		p.conn.RetryReadLine(7) // Just to consume their next | fin:sync
		return p.conn.Err()
	} else {
		p.conn.TransmitStringLine("next")
		resp := p.conn.RetryReadStringLine(7)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		if resp == "fin:sync" {
			return nil
		} else if resp != "next" {
			return fmt.Errorf("expected 'next' | 'fin:sync', received %s", resp)
		}
	}
	if theirWork.Lt(ourWork) {
		// Send a sync
		p.conn.TransmitStringLine("sync-send")
		p.conn.ConsumeExpected("sync-recv")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return p.handleSendSync()
	} else {
		// Receive a sync
		p.conn.TransmitStringLine("sync-recv")
		p.conn.ConsumeExpected("sync-send")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return p.handleReceiveSync()
	}
}

// Send a chain sync.
func (p *Peer) handleSendSync() error {
	// Find last common anceskern.HashT peer
	neededBlockIds := make([]kern.HashT, 0)
	lcaId := p.head
	p.conn.TransmitHashLine(lcaId)
	resp := p.conn.RetryReadStringLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	for resp == "next" {
		neededBlockIds = append(neededBlockIds, lcaId)
		lcaId = p.inv.GetBlockParentId(lcaId)
		p.conn.TransmitHashLine(lcaId)
		resp = p.conn.RetryReadStringLine(7)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	if resp != "recognized" {
		return fmt.Errorf("expected 'recognized', received %s", resp)
	}
	if len(neededBlockIds) == 0 {
		return fmt.Errorf("peer does not need upgrade, sync should not have run")
	}
	// Send blocks to peer
	for _, blockId := range neededBlockIds {
		p.conn.TransmitBlockHeader(p.inv.GetBlock(blockId))
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	// Check if peer verified the chain's work
	resp = p.conn.RetryReadStringLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	if resp == "reject" {
		return fmt.Errorf("peer does not actually want sync")
	} else if resp != "next" {
		return fmt.Errorf("expected 'next', received %s", resp)
	}
	// Receive entity peer wants and reply with it, until peer sends zero value hash
	entityId := p.conn.RetryReadHashLine(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	for !entityId.EqZero() {
		if p.inv.HasMerkle(entityId) {
			p.conn.TransmitStringLine("merkle")
			p.conn.TransmitMerkle(p.inv.GetMerkle(entityId))
		} else if p.inv.HasTx(entityId) {
			p.conn.TransmitStringLine("tx")
			p.conn.TransmitTx(p.inv.GetTx(entityId))
		} else {
			return fmt.Errorf("peer requested unknown entity %s", entityId)
		}
		// Receive next entity
		entityId = p.conn.RetryReadHashLine(7)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	return nil
}

// Receive a chain sync.
func (p *Peer) handleReceiveSync() error {
	// Find last common anceskern.HashT peer
	neededBlockIds := make([]kern.HashT, 0)
	newHead := p.conn.RetryReadHashLine(7)
	lcaId := newHead
	for !p.inv.HasBlock(lcaId) {
		neededBlockIds = append(neededBlockIds, lcaId)
		p.conn.TransmitStringLine("next")
		lcaId = p.conn.RetryReadHashLine(7)
	}
	p.conn.TransmitStringLine("recognized")
	if len(neededBlockIds) == 0 {
		return fmt.Errorf("we do not need upgrade, sync should not have run")
	}
	// Receive blocks frokern.HashT
	blockMap := make(map[kern.HashT]kern.Block)
	for _, blockId := range neededBlockIds {
		blockMap[blockId] = p.conn.RetryReadBlockHeader(7, blockId)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	// Do light verification (most importantly proof-of-work)
	// Contains some ddos attempts to this loop and away from manager
	if err := p.quickVerifyChain(newHead, lcaId, neededBlockIds, blockMap); err != nil {
		p.conn.TransmitStringLine("reject")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		return err
	}
	p.conn.TransmitStringLine("next")
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	// Until our inv / local inv is complete, request entities, and add to output
	newBlocks := make([]kern.Block, 0)
	newMerkles := make([]kern.MerkleNode, 0)
	newTxs := make([]kern.Tx, 0)
	entityQueue := util.NewQueue[kern.HashT]()
	newEntitySet := util.NewSet[kern.HashT]()
	for _, blockId := range neededBlockIds {
		merkle := blockMap[blockId].MerkleRoot
		if !p.inv.HasMerkle(merkle) {
			entityQueue.Push(merkle)
		}
		newBlocks = util.Prepend(newBlocks, blockMap[blockId])
	}
	for entityQueue.Size() > 0 {
		entityId, _ := entityQueue.Pop()
		if p.inv.HasEntity(entityId) || newEntitySet.Includes(entityId) {
			continue
		}
		p.conn.TransmitHashLine(entityId)
		entityType := p.conn.RetryReadStringLine(7)
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		if entityType == "merkle" {
			merkle := p.conn.RetryReadMerkle(7, entityId)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			newMerkles = util.Prepend(newMerkles, merkle)
			entityQueue.Push(merkle.LChild, merkle.RChild)
		} else if entityType == "tx" {
			tx := p.conn.RetryReadTx(7, entityId)
			if p.conn.HasErr() {
				return p.conn.Err()
			}
			newTxs = util.Prepend(newTxs, tx)
		} else {
			return fmt.Errorf("unrecognized entity type: %s", entityType)
		}
		if p.conn.HasErr() {
			return p.conn.Err()
		}
		newEntitySet.Add(entityId)
	}
	p.conn.TransmitHashLine(kern.HashT{})
	// Send the event to the manager
	p.mainHandler.HandleInboundSync(newHead, newBlocks, newMerkles, newTxs)
	return nil
}

// Verify a new chain's continuity, expected endpoints, and proof-of-work.
// Leaves heavier and state-based verification to manager, this just helps prevent ddos.
func (p *Peer) quickVerifyChain(
	newHead kern.HashT,
	lcaId kern.HashT,
	neededBlockIds []kern.HashT,
	blockMap map[kern.HashT]kern.Block,
) error {
	// Verify chain has expected head
	if neededBlockIds[0] != newHead {
		return fmt.Errorf("received chain does not have expected head")
	}
	// Verify chain continuous
	for i := 0; i < len(neededBlockIds)-1; i++ {
		if blockMap[neededBlockIds[i]].PrevBlockId != neededBlockIds[i+1] {
			return fmt.Errorf("received chain not continuous")
		}
	}
	// Verify chain attaches where expected
	if blockMap[neededBlockIds[len(neededBlockIds)-1]].PrevBlockId != lcaId {
		return fmt.Errorf("received chain does not attach to ours as expected")
	}
	// Verify claimed total work beats ours
	newWork := p.inv.GetBlockTotalWork(lcaId)
	for _, blockId := range neededBlockIds {
		newWork = newWork.WorkAppendTarget(blockMap[blockId].Difficulty)
	}
	if !p.inv.GetBlockTotalWork(p.head).Lt(newWork) {
		return fmt.Errorf("received chain is not actually higher work than ours")
	}
	// Verify each block actually beats claimed difficulty
	for _, blockId := range neededBlockIds {
		if !blockId.Lt(blockMap[blockId].Difficulty) {
			return fmt.Errorf("received block does not actually beat difficulty")
		}
	}
	return nil
}

// Handle the receipt of new addresses from the peer.
func (p *Peer) handleReceiveAddrs() error {
	numAddrs := p.conn.RetryReadUint64Line(7)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	addrs := make([]string, numAddrs)
	for i := range addrs {
		addrs[i] = p.conn.RetryReadStringLine(7)
	}
	p.conn.ConsumeExpected("fin:addrs")
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	p.mainHandler.HandlePeersReceived(addrs)
	return nil
}

// Handle the receipt of a peers wanted message.
func (p *Peer) handleReceivePeersWanted() error {
	p.mainHandler.HandlePeersWanted(p.Info.RuntimeID)
	return nil
}

// Handle the receipt of a new tx.
func (p *Peer) handleReceiveTx() error {
	txId := p.conn.RetryReadHashLine(7)
	if p.inv.HasTx(txId) {
		p.conn.TransmitStringLine("fin:tx")
		return p.conn.Err()
	}
	p.conn.TransmitStringLine("next")
	tx := p.conn.RetryReadTx(7, txId)
	if p.conn.HasErr() {
		return p.conn.Err()
	}
	p.mainHandler.HandleNewTx(tx)
	return nil
}
