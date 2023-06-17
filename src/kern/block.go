package kern

import (
	"fmt"
	"math"

	"github.com/levilutz/basiccoin/src/util"
)

type MerkleNode struct {
	LChild HashT
	RChild HashT
}

func (node MerkleNode) Hash() HashT {
	return DHashVarious(node.LChild, node.RChild)
}

type Block struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Noise       HashT
	Nonce       uint64
}

func (b Block) Hash() HashT {
	return DHashVarious(b.PrevBlockId, b.MerkleRoot, b.Difficulty, b.Noise, b.Nonce)
}

// Verify that the claimed proof of work is valid.
func (b Block) VerifyProofOfWork() error {
	if !b.Hash().Lt(b.Difficulty) {
		return fmt.Errorf("failed to beat claimed target")
	}
	return nil
}

// The maximum number of txs that could theoretically be in a block, including coinbase.
func BlockMaxTxs() uint64 {
	standardTxSpace := util.Constants.MaxBlockVSize - CoinbaseVSize()
	// +1 to "round up"
	maxStandardTxs := standardTxSpace/MinNonCoinbaseVSize() + 1
	// +1 to re-include coinbase tx
	return maxStandardTxs + 1
}

// The (overestimated) max possible size of any block's merkle tree, including tx leafs.
func MerkleTreeMaxSize() uint64 {
	// Actual tree size <= floor(leafs * 20 / 9)
	return uint64(float64(BlockMaxTxs()) * 20.0 / 9.0)
}

// Construct a merkle tree from a list of txIds.
// Returns merkle nodes by hash, and order they should be inserted into inv.
// Last id in list is the root.
func MerkleFromTxIds(txIds []HashT) (map[HashT]MerkleNode, []HashT) {
	if len(txIds) == 0 {
		panic("need at least one tx to generate tree")
	}
	// Special case - if 1 tx, we still want a tree above it
	if len(txIds) == 1 {
		node := MerkleNode{LChild: txIds[0], RChild: txIds[0]}
		nodeId := node.Hash()
		return map[HashT]MerkleNode{nodeId: node}, []HashT{nodeId}
	}
	// Normal case - generate layers one at a time
	outMap := make(map[HashT]MerkleNode)
	// Number of tree layers, _including_ txs at bottom
	numLayers := int(math.Ceil(math.Log2(float64(len(txIds))))) + 1
	// Initialize tree layers
	outLayers := make([][]HashT, numLayers)
	for i := range outLayers {
		outLayers[i] = make([]HashT, 0)
	}
	// Populate lowest layer of tree with txIds
	outLayers[0] = append(outLayers[0], txIds...)
	// Populate higher layers of tree successively (skip lowest)
	for l := 1; l < numLayers; l++ {
		lastLayer := outLayers[l-1]
		// Insert pairs of elements
		for i := 0; i < len(lastLayer)/2; i++ {
			node := MerkleNode{
				LChild: lastLayer[2*i],
				RChild: lastLayer[2*i+1],
			}
			nodeId := node.Hash()
			outMap[nodeId] = node
			outLayers[l] = append(outLayers[l], nodeId)
		}
		// Insert trailing odd element if exists
		if len(lastLayer)%2 == 1 {
			node := MerkleNode{
				LChild: lastLayer[len(lastLayer)-1],
				RChild: lastLayer[len(lastLayer)-1],
			}
			nodeId := node.Hash()
			outMap[nodeId] = node
			outLayers[l] = append(outLayers[l], nodeId)
		}
	}
	return outMap, util.FlattenLists(outLayers[1:])
}
