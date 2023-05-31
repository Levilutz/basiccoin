package peer

import (
	"fmt"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/events"
)

// Handle a new block exchange, returning event for manager.
func handleReceiveNewBlockExchange(
	pc *PeerConn, inv db.InvReader,
) (event events.CandidateLedgerUpgradeMainEvent, err error) {
	var recId db.HashT
	var ok bool
	// Exchange init
	topBlockIdStr := pc.RetryReadStringLine(7)
	if pc.HasErr() {
		return event, pc.Err()
	}
	topBlockId, err := db.StringToHash(topBlockIdStr)
	if err != nil {
		return event, err
	}
	if inv.HasBlock(topBlockId) {
		pc.TransmitStringLine("fin:new-block")
		if pc.HasErr() {
			return event, pc.Err()
		}
		return event, fmt.Errorf("block id known")
	}
	// Exchange block ids
	pc.TransmitStringLine("next-blocks")
	if pc.HasErr() {
		return event, pc.Err()
	}
	neededBlockIds := []db.HashT{
		topBlockId,
	}
	for {
		newBlockIds, err := ReceiveBlockIdsMessage(pc)
		if err != nil {
			return event, err
		}
		recId, ok = inv.HasAnyBlock(newBlockIds.BlockIds)
		// Add to list of needed block ids, until we hit the one we recognize
		for _, blockId := range newBlockIds.BlockIds {
			if ok && recId == blockId {
				break
			}
			neededBlockIds = append(neededBlockIds, blockId)
		}
		// Stop the initiator from sending more block ids
		if ok {
			pc.TransmitStringLine("recognized")
			pc.TransmitMessage(BlockIdsMessage{BlockIds: []db.HashT{recId}})
			if pc.HasErr() {
				return event, pc.Err()
			}
			break
		}
	}
	// Exchange block headers
	blocks := make(map[db.HashT]db.Block, len(neededBlockIds))
	for _, expectedBlockId := range neededBlockIds {
		block, err := ReceiveBlockHeaderMessage(pc)
		if err != nil {
			return event, err
		}
		blockId := block.Block.Hash()
		if blockId != expectedBlockId {
			return event, fmt.Errorf(
				"mismatched block %x != %x", blockId, expectedBlockId,
			)
		}
		blocks[blockId] = block.Block
	}
	// Verify valid chain and proof of work (contains some ddos attacks to peer thread)
	for i := 0; i < len(neededBlockIds)-1; i++ {
		if blocks[neededBlockIds[i]].PrevBlockId != neededBlockIds[i+1] {
			return event, fmt.Errorf("new block parent mismatched")
		}
	}
	lastBlockId := neededBlockIds[len(neededBlockIds)-1]
	if blocks[lastBlockId].PrevBlockId != recId {
		return event, fmt.Errorf("last block does not attach to our chain")
	}
	// TODO: Verify proof of work beats ours (requires we know head)
	// Check if higher work than existing head (from branch point)
	// If not, terminate exchange
	// Exchange merkles and txs
	return event, nil
}

func handleTransmitNewBlockExchange(
	blockId db.HashT, pc *PeerConn, inv db.InvReader,
) error {
	// Exchange init
	pc.TransmitStringLine(fmt.Sprintf("%x", blockId))
	resp := pc.RetryReadStringLine(7)
	if pc.HasErr() {
		return pc.Err()
	}
	if resp == "fin:new-block" {
		return nil
	} else if resp != "next-blocks" {
		return fmt.Errorf("unexpected response: %s", resp)
	}
	// Exchange block ids
	for resp == "next-blocks" {
		nextBlocks := inv.GetBlockAncestors(blockId, 20)
		pc.TransmitMessage(BlockIdsMessage{BlockIds: nextBlocks})
		resp = pc.RetryReadStringLine(7)
		if pc.HasErr() {
			return pc.Err()
		}
	}
	if resp != "recognized" {
		return fmt.Errorf("unexpected response: %s", resp)
	}
	_, err := ReceiveBlockIdsMessage(pc)
	if err != nil {
		return err
	}
	// Exchange block headers
	// Exchange merkles and txs
	return nil
}
