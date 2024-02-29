package commitments

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
)

const (
	// BridgeCommitmentBlocksLimit is the limit to the number of blocks we can generate a bridge commitment for.
	BridgeCommitmentBlocksLimit = 1000
)

func (dcs DataCommitmentsServer) validateBridgeCommitmentRange(
	clientCtx client.Context, start uint64, end uint64,
) error {
	if start == 0 {
		return fmt.Errorf("the first block is 0")
	}
	heightsRange := end - start
	if heightsRange > uint64(BridgeCommitmentBlocksLimit) {
		return fmt.Errorf("the query exceeds the limit of allowed blocks %d", BridgeCommitmentBlocksLimit)
	}
	if heightsRange == 0 {
		return fmt.Errorf("cannot create the bridge commitments for an empty set of blocks")
	}
	if start >= end {
		return fmt.Errorf("last block is smaller than first block")
	}
	// The bridge commitment range is end exclusive
	height, err := getBlockHeight(clientCtx)
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
func (dcs DataCommitmentsServer) validateBridgeCommitmentInclusionProofRequest(
	clientCtx client.Context, height uint64, start uint64, end uint64,
) error {
	err := dcs.validateBridgeCommitmentRange(clientCtx, start, end)
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
