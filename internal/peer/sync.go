package peer

import "fmt"

var syncCmd = "chain-sync"

// Handle a sync, inbound or outbound.
func (p *Peer) handleSync() error {
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
	} else {
		// Receive a sync
		p.conn.WriteString("sync:recv")
		p.conn.ReadStringExpected("sync:send")
		if p.conn.HasErr() {
			return p.conn.Err()
		}
	}
	return nil
}
