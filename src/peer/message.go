package peer

import (
	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

type PeerMessage interface {
	Transmit(pc *PeerConn) error
}

// Receive base64(json(message)) from a single line
func receiveStandardMessage[R PeerMessage](pc *PeerConn) (R, error) {
	// Cannot be method until golang allows type params on methods
	var content R
	data := pc.RetryReadLine(7)
	if err := pc.Err(); err != nil {
		return content, err
	}
	return util.UnJsonB64[R](data)
}

// HelloMessage

type HelloMessage struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
	Addr      string `json:"addr"`
}

// Construct a HelloMessage
func NewHelloMessage() HelloMessage {
	addr := ""
	if util.Constants.Listen {
		addr = util.Constants.LocalAddr
	}
	return HelloMessage{
		RuntimeID: util.Constants.RuntimeID,
		Version:   util.Constants.Version,
		Addr:      addr,
	}
}

// Receive a HelloMessage from the channel
func ReceiveHelloMessage(pc *PeerConn) (HelloMessage, error) {
	return receiveStandardMessage[HelloMessage](pc)
}

// Transmit a HelloMessage over the channel
func (msg HelloMessage) Transmit(pc *PeerConn) error {
	data, err := util.JsonB64(msg)
	if err != nil {
		return err
	}
	pc.TransmitLine(data)
	return pc.Err()
}

// AddrsMessage

type AddrsMessage struct {
	PeerAddrs []string `json:"peerAddrs"`
}

// Construct an AddrsMessage
func ReceiveAddrsMessage(pc *PeerConn) (AddrsMessage, error) {
	numAddrs := pc.RetryReadIntLine(7)
	if err := pc.Err(); err != nil {
		return AddrsMessage{}, err
	}
	addrs := make([]string, numAddrs)
	for i := 0; i < numAddrs; i++ {
		addrs[i] = pc.RetryReadStringLine(7)
	}
	pc.ConsumeExpected("fin:addrs")
	return AddrsMessage{
		PeerAddrs: addrs,
	}, pc.Err()
}

// Transmit an AddrsMessage over the channel
func (msg AddrsMessage) Transmit(pc *PeerConn) error {
	pc.TransmitIntLine(len(msg.PeerAddrs))
	for _, addr := range msg.PeerAddrs {
		if addr == "fin:addrs" {
			continue
		}
		pc.TransmitStringLine(addr)
	}
	pc.TransmitStringLine("fin:addrs")
	return pc.Err()
}

// BlockIdsMessage

type BlockIdsMessage struct {
	BlockIds []db.HashT
}

// Construct a BlockIdsMessage
func ReceiveBlockIdsMessage(pc *PeerConn) (BlockIdsMessage, error) {
	pc.ConsumeExpected("block-ids")
	numBlockIds := pc.RetryReadIntLine(7)
	line := pc.RetryReadStringLine(7)
	if err := pc.Err(); err != nil {
		return BlockIdsMessage{}, err
	}
	hashes, err := db.StringToHashes(line, numBlockIds)
	if err != nil {
		return BlockIdsMessage{}, err
	}
	pc.ConsumeExpected("fin:block-ids")
	if err := pc.Err(); err != nil {
		return BlockIdsMessage{}, err
	}
	return BlockIdsMessage{
		BlockIds: hashes,
	}, nil
}

// Transmit a BlockIdsMessage over the channel
func (msg BlockIdsMessage) Transmit(pc *PeerConn) error {
	pc.TransmitStringLine("block-ids")
	pc.TransmitIntLine(len(msg.BlockIds))
	pc.TransmitStringLine(db.HashesToString(msg.BlockIds))
	pc.TransmitStringLine("fin:block-ids")
	return pc.Err()
}

// BlockHeaderMessage

type BlockHeaderMessage struct {
	Block db.Block
}

// Construct a BlockHeaderMessage
func ReceiveBlockHeaderMessage(pc *PeerConn) (BlockHeaderMessage, error) {
	hashesLine := pc.RetryReadStringLine(7)
	nonce := pc.RetryReadUint64Line(7)
	if err := pc.Err(); err != nil {
		return BlockHeaderMessage{}, err
	}
	hashes, err := db.StringToHashes(hashesLine, 4)
	if err != nil {
		return BlockHeaderMessage{}, err
	}
	return BlockHeaderMessage{
		Block: db.Block{
			PrevBlockId: hashes[0],
			MerkleRoot:  hashes[1],
			Difficulty:  hashes[2],
			Noise:       hashes[3],
			Nonce:       nonce,
		},
	}, nil
}

// Transmit a BlockHeaderMessage over the channel.
func (msg BlockHeaderMessage) Transmit(pc *PeerConn) error {
	pc.TransmitStringLine(db.HashesToString([]db.HashT{
		msg.Block.PrevBlockId,
		msg.Block.MerkleRoot,
		msg.Block.Difficulty,
		msg.Block.Noise,
	}))
	pc.TransmitUint64Line(msg.Block.Nonce)
	return pc.Err()
}
