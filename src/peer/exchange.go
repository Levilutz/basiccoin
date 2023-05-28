package peer

import (
	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/events"
)

// Handle a new block exchange, returning event for manager.
func handleNewBlockExchange(
	pc *PeerConn, invReader db.InvReader,
) (events.CandidateLedgerUpgradeMainEvent, error) {
	// Exchange block ids
	// Exchange block headers
	// Check if higher work than existing head (from branch point)
	// If not, terminate exchange
	// Exchange merkles and txs
	return events.CandidateLedgerUpgradeMainEvent{}, nil
}
