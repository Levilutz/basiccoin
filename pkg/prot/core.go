package prot

import "github.com/levilutz/basiccoin/pkg/core"

// Read a HashT from the conn.
func (c *Conn) ReadHashT() core.HashT {
	if c.err != nil {
		return core.HashT{}
	}
	raw := c.readRawTimeout(32, defaultTimeout)
	if c.err != nil {
		return core.HashT{}
	}
	return core.NewHashTFromBytes(raw)
}

// Write a HashT to the conn.
func (c *Conn) WriteHashT(data core.HashT) {
	if c.err != nil {
		return
	}
	actual := data.Data()
	c.writeRawTimeout(actual[:], defaultTimeout)
}

// Read a Block from the conn.
func (c *Conn) ReadBlock() core.Block {
	if c.err != nil {
		return core.Block{}
	}
	return core.Block{
		PrevBlockId: c.ReadHashT(),
		MerkleRoot:  c.ReadHashT(),
		Target:      c.ReadHashT(),
		Noise:       c.ReadHashT(),
		Nonce:       c.ReadUint64(),
		MinedTime:   c.ReadUint64(),
	}
}

// Write a Block to the conn.
func (c *Conn) WriteBlock(data core.Block) {
	if c.err != nil {
		return
	}
	c.WriteHashT(data.PrevBlockId)
	c.WriteHashT(data.MerkleRoot)
	c.WriteHashT(data.Target)
	c.WriteHashT(data.Noise)
	c.WriteUint64(data.Nonce)
	c.WriteUint64(data.MinedTime)
}

// Read a Merkle Node from the conn.
func (c *Conn) ReadMerkle() core.MerkleNode {
	if c.err != nil {
		return core.MerkleNode{}
	}
	return core.MerkleNode{
		LChild: c.ReadHashT(),
		RChild: c.ReadHashT(),
	}
}

// Write a Merkle Node to the conn.
func (c *Conn) WriteMerkle(data core.MerkleNode) {
	if c.err != nil {
		return
	}
	c.WriteHashT(data.LChild)
	c.WriteHashT(data.RChild)
}

// Read a Tx from the conn.
func (c *Conn) ReadTx() core.Tx {
	if c.err != nil {
		return core.Tx{}
	}
	tx := core.Tx{
		IsCoinbase: c.ReadBool(),
		MinBlock:   c.ReadUint64(),
		Inputs:     make([]core.TxIn, c.ReadUint64()),
		Outputs:    make([]core.TxOut, c.ReadUint64()),
	}
	if c.err != nil {
		return core.Tx{}
	}
	for i := range tx.Inputs {
		tx.Inputs[i] = core.TxIn{
			Utxo: core.Utxo{
				TxId:  c.ReadHashT(),
				Ind:   c.ReadUint64(),
				Value: c.ReadUint64(),
			},
			PublicKey: c.Read(),
			Signature: c.Read(),
		}
	}
	for i := range tx.Outputs {
		tx.Outputs[i] = core.TxOut{
			Value:         c.ReadUint64(),
			PublicKeyHash: c.ReadHashT(),
		}
	}
	return tx
}

// Write a Tx to the conn.
func (c *Conn) WriteTx(data core.Tx) {
	if c.err != nil {
		return
	}
	c.WriteBool(data.IsCoinbase)
	c.WriteUint64(data.MinBlock)
	c.WriteUint64(uint64(len(data.Inputs)))
	c.WriteUint64(uint64(len(data.Outputs)))
	for _, txi := range data.Inputs {
		c.WriteHashT(txi.Utxo.TxId)
		c.WriteUint64(txi.Utxo.Ind)
		c.WriteUint64(txi.Utxo.Value)
		c.Write(txi.PublicKey)
		c.Write(txi.Signature)
	}
	for _, txo := range data.Outputs {
		c.WriteUint64(txo.Value)
		c.WriteHashT(txo.PublicKeyHash)
	}
}
