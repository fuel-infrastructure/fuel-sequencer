package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
)

const (
	// BridgeCommitmentBlocksLimit is the limit to the number of blocks we can generate a bridge commitment for.
	// This limits the ZK prover time required to compute.
	// Inspiration: https://github.com/celestiaorg/celestia-core/blob/v1.35.0-tm-v0.34.29/pkg/consts/consts.go#L43-L44
	BridgeCommitmentBlocksLimit = 4096
)

var (
	ErrZeroStart = errors.New("the first block is 0")
)

// validateBridgeCommitmentRange runs basic checks on the range of heights
// that will be used to generate bridge commitments from successive blocks.
func validateBridgeCommitmentRange(ctx context.Context, clientCtx client.Context, start, end uint64) error {
	if start == 0 {
		return ErrZeroStart
	}
	if start > end {
		return fmt.Errorf("last block is smaller than first block")
	}
	heightsRange := end - start
	if heightsRange == 0 {
		return fmt.Errorf("cannot create the bridge commitments for an empty set of blocks")
	}
	if heightsRange > uint64(BridgeCommitmentBlocksLimit) {
		return fmt.Errorf("the query exceeds the limit of allowed blocks %d", BridgeCommitmentBlocksLimit)
	}
	// The bridge commitment range is end exclusive.
	height, err := getLatestBlockHeight(ctx, clientCtx)
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
func validateBridgeCommitmentInclusionProofRequest(
	ctx context.Context, clientCtx client.Context, height, start, end uint64,
) error {
	err := validateBridgeCommitmentRange(ctx, clientCtx, start, end)
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
