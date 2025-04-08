package service

import (
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

// encodeBridgeCommitment takes a height and a last result hash, and returns the equivalent of
// `abi.encode(...)` in Ethereum. To match `abi.encode(...)`, the height is padded to 32 bytes.
func encodeBridgeCommitment(leaves []types.BridgeCommitmentLeaf) ([][]byte, error) {

	encodedLeaves := make([][]byte, 0, len(leaves))
	for _, leaf := range leaves {

		// Pad to match `abi.encode` on Ethereum.
		paddedHeight, err := utils.To32PaddedHexBytes(leaf.Height)
		if err != nil {
			return nil, err
		}

		encodedLeaf := append(paddedHeight, leaf.LastResultsHash...)
		encodedLeaves = append(encodedLeaves, encodedLeaf)
	}

	return encodedLeaves, nil
}
