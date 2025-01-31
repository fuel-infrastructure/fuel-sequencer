package service

import (
	"context"
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

// validateBridgeCommitmentRange runs basic checks on the range of heights
// that will be used to generate bridge commitments from successive blocks.
func (q queryServer) validateBridgeCommitmentRange(ctx context.Context, start, end uint64) error {
	if start == 0 {
		return types.ErrZeroStart
	}
	if start > end {
		return fmt.Errorf("last block is smaller than first block")
	}
	heightsRange := end - start
	if heightsRange == 0 {
		return fmt.Errorf("cannot create the bridge commitments for an empty set of blocks")
	}
	if heightsRange > q.maxQueryRange {
		return fmt.Errorf("the query exceeds the maximum block range %d", q.maxQueryRange)
	}
	// The bridge commitment range is end exclusive.
	height, err := getLatestBlockHeight(ctx, q.node)
	if err != nil {
		return err
	}
	if end > uint64(height)+1 {
		return fmt.Errorf(
			"end block %d is higher than current chain height %d",
			end,
			height,
		)
	}
	return nil
}

// validateBridgeCommitmentInclusionProofRequest validates the request to generate a bridge commitment
// inclusion proof.
func (q queryServer) validateBridgeCommitmentInclusionProofRequest(
	ctx context.Context, height, start, end uint64,
) error {
	err := q.validateBridgeCommitmentRange(ctx, start, end)
	if err != nil {
		return err
	}
	if height < start || height >= end {
		return fmt.Errorf(
			"height %d should be in the end exclusive interval first_block %d last_block %d",
			height,
			start,
			end,
		)
	}
	return nil
}
