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
	// Exchange init
	blockIdStr := pc.RetryReadStringLine(7)
	if err := pc.Err(); err != nil {
		return event, err
	}
	blockId, err := db.StringToHash(blockIdStr)
	if err != nil {
		return event, err
	}
	if _, ok := inv.LoadBlock(blockId); ok {
		pc.TransmitStringLine("fin:new-block")
		return event, fmt.Errorf("block id known")
	}
	// Exchange block ids
	pc.TransmitStringLine("next-blocks")
	for {
		newBlockIds, err := ReceiveBlockIdsMessage(pc)
		if err != nil {
			return event, err
		}
		if recId, ok := inv.AnyBlockIdsKnown(newBlockIds.BlockIds); ok {
			pc.TransmitStringLine("recognized")
			pc.TransmitStringLine(fmt.Sprintf("%x", recId))
			break
		}
	}
	// Exchange block headers
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
	if err := pc.Err(); err != nil {
		return err
	}
	if resp == "fin:new-block" {
		return nil
	} else if resp != "next-blocks" {
		return fmt.Errorf("unexpected response: %s", resp)
	}
	// Exchange block ids
	for resp == "next-blocks" {
		nextBlocks, err := inv.GetBlockHeritage(blockId, 20)
		if err != nil {
			return err
		}
		pc.TransmitMessage(BlockIdsMessage{BlockIds: nextBlocks})
		resp = pc.RetryReadStringLine(7)
		if err := pc.Err(); err != nil {
			return err
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
