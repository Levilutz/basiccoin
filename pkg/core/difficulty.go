package core

import (
	"fmt"
	"math/big"
)

// Subset of Inv methods required to compute a next target.
type InvTargeter interface {
	GetBlock(blockId HashT) Block
	GetBlockHeight(blockId HashT) uint64
	GetBlockAncestors(blockId HashT, maxLen int) []HashT
	GetBlockSpecificAncestor(blockId HashT, depth int) HashT
}

// Compute actual and desired time to compute the blocks in this period so far.
func ExpectedTargetAdjustment(params Params, inv InvTargeter, head HashT) (uint64, uint64, error) {
	headHeight := inv.GetBlockHeight(head)

	// Check there exists data to compute adjustment from
	if headHeight < 2 || headHeight%params.DifficultyPeriod == 0 {
		return 0, 0, fmt.Errorf("too soon to compute target adjustment")
	}

	// Get height of first block in this period and number of hops to it
	var firstHeight uint64
	if headHeight < params.DifficultyPeriod {
		firstHeight = 1
	} else {
		firstHeight = (headHeight / params.DifficultyPeriod) * params.DifficultyPeriod
	}
	numBlocksMeasured := headHeight - firstHeight

	// Compute actual vs desired time to mine these blocks
	headMinedTime := inv.GetBlock(head).MinedTime
	periodFirstId := inv.GetBlockSpecificAncestor(head, int(numBlocksMeasured))
	firstMinedTime := inv.GetBlock(periodFirstId).MinedTime
	if firstMinedTime >= headMinedTime {
		// Should be protected from happening on last block by VerifyBlock
		return 0, 0, fmt.Errorf("cannot adjust target - head was mined 'before' first block")
	}
	actualTime := headMinedTime - firstMinedTime
	desiredTime := params.BlockTargetTime * numBlocksMeasured

	// Return each component rather than deciding which way the division should go here.
	return actualTime, desiredTime, nil
}

func NextTarget(params Params, inv InvTargeter, prevBlockId HashT) HashT {
	// Special case - first block gets original target
	if prevBlockId.EqZero() {
		return params.OriginalTarget
	}

	// Most common case - no adjustment
	if (inv.GetBlockHeight(prevBlockId)+1)%params.DifficultyPeriod != 0 {
		return inv.GetBlock(prevBlockId).Target
	}

	// Actually adjust target
	actualTime, desiredTime, err := ExpectedTargetAdjustment(params, inv, prevBlockId)
	if err != nil {
		// Should be guarded against by VerifyBlock and params.verify
		panic(err)
	}

	// target = prevTarget * actualTime / desiredTime
	prevTarget := inv.GetBlock(prevBlockId).Target
	targetInt := prevTarget.BigInt()
	actualInt := &big.Int{}
	actualInt.SetUint64(actualTime)
	desiredInt := &big.Int{}
	desiredInt.SetUint64(desiredTime)
	targetInt.Mul(targetInt, actualInt)
	targetInt.Div(targetInt, desiredInt)

	// target = min(target, 2^256-1)
	if targetInt.Cmp(bigInt2_256()) == 1 {
		targetInt = bigInt2_256()
	}
	target := NewHashTFromBigInt(targetInt)

	// Prevent target from reducing more than a factor of 4
	if minNext := prevTarget.MinNextTarget(); target.Lt(minNext) {
		target = minNext
	}

	// Prevent target from increasing more than a factor of 4
	if maxNext := prevTarget.MaxNextTarget(params); maxNext.Lt(target) {
		target = maxNext
	}

	adjustmentPct := 100.0 * float64(actualTime) / float64(desiredTime)
	fmt.Printf("adjusting target to %f pct of previous - %s\n", adjustmentPct, target)
	return target
}
