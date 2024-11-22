package service

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/crypto/merkle"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

type queryServer struct {
	clientCtx         client.Context
	interfaceRegistry codectypes.InterfaceRegistry
	maxQueryRange     uint64
}

func NewQueryServer(
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
	maxQueryRange uint64,
) types.QueryServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
		maxQueryRange:     maxQueryRange,
	}
}

// BridgeCommitment collects the transactions results roots over a provided ordered range of blocks,
// and then creates a new merkle root. The range is end exclusive.
func (q queryServer) BridgeCommitment(
	ctx context.Context, req *types.QueryBridgeCommitmentRequest,
) (*types.QueryBridgeCommitmentResponse, error) {
	defer telemetry.MeasureSince(telemetry.Now(), "sequencer", "query", "bridge", "commitment")

	err := q.validateBridgeCommitmentRange(ctx, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	// Fetch data.
	leaves, err := fetchBridgeCommitmentLeaves(ctx, q.clientCtx, req.Start, req.End)
	if err != nil {
		return nil, err
	}
	// Encode data to match solidity `abi.encode`.
	encodedLeaves, err := encodeBridgeCommitment(leaves)
	if err != nil {
		return nil, err
	}
	root := merkle.HashFromByteSlices(encodedLeaves)

	return &types.QueryBridgeCommitmentResponse{
		BridgeCommitment: root,
	}, nil
}

// BridgeCommitmentInclusionProof creates two inclusion proofs to verify that a transaction response is
// included in a BridgeCommitment. Users can also verify any data in the transactions responses for data
// availability. Users need to provide the height and the transaction index that the inclusion proof is
// for. They also need to provide the indexes of the start and end blocks for which the BridgeCommitment
// merkle root is constructed from. The range for BridgeCommitment is end exclusive.
func (q queryServer) BridgeCommitmentInclusionProof(
	ctx context.Context, req *types.QueryBridgeCommitmentInclusionProofRequest,
) (*types.QueryBridgeCommitmentInclusionProofResponse, error) {
	defer telemetry.MeasureSince(telemetry.Now(), "sequencer", "query", "bridge", "commitment", "inclusion", "proof")

	err := q.validateBridgeCommitmentInclusionProofRequest(ctx, uint64(req.Height), req.Start, req.End)
	if err != nil {
		return nil, err
	}

	// Fetch data.
	leaves, err := fetchBridgeCommitmentLeaves(ctx, q.clientCtx, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	// Get the relevant index within the specified range of blocks.
	// e.g. for block 1500 in 1000 to 2000, the index is 1500-1000 = 500
	blockIndex := req.Height - int64(leaves[0].Height)

	// Encode data to match solidity `abi.encode`.
	encodedLeaves, err := encodeBridgeCommitment(leaves)
	if err != nil {
		return nil, err
	}
	// Get the relevant proof of the BridgeCommitment leaves.
	_, proofs := merkle.ProofsFromByteSlices(encodedLeaves)
	bcProof := proofs[blockIndex]

	// Get the relevant BridgeCommitmentLeaf.
	bcLeaf := leaves[blockIndex]

	// Load the transactions that composed the LastResultsHash at height.
	txResultHeight := req.Height - 1
	finalizeBlockResponse, err := getBlockResults(ctx, q.clientCtx, &txResultHeight)
	if err != nil {
		return nil, err
	}

	// If there are no transactions in the block it is not possible to generate an inclusion proof for a transaction
	// response. However, we can still return the proof showing that the block was included in the bridge commitment.
	if len(finalizeBlockResponse.TxsResults) == 0 && int(req.TxIndex) == 0 {
		return &types.QueryBridgeCommitmentInclusionProofResponse{
			BridgeCommitmentProof: *types.NewBinaryMerkleProof(*bcProof),
			BridgeCommitmentLeaf:  bcLeaf,
		}, nil
	}

	// Sanity check.
	if int(req.TxIndex) >= len(finalizeBlockResponse.TxsResults) {
		return nil, fmt.Errorf("transaction index too high %d", req.TxIndex)
	}

	// Remove non-deterministic fields from ExecTxResult responses to match LastResultsHash from the
	// header computation. Ref: https://github.com/cometbft/cometbft/blob/v0.38.5/state/store.go#L412
	deterministicTxResults := cmtypes.NewResults(finalizeBlockResponse.TxsResults)
	// Get the merkle proof for this transaction.
	txMerkleProof := deterministicTxResults.ProveResult(int(req.TxIndex))
	// Get the marshalled transaction result
	txResultMarshalled, err := deterministicTxResults[req.TxIndex].Marshal()
	if err != nil {
		return nil, err
	}

	return &types.QueryBridgeCommitmentInclusionProofResponse{
		BridgeCommitmentProof: *types.NewBinaryMerkleProof(*bcProof),
		LastResultsProof:      types.NewBinaryMerkleProof(txMerkleProof),
		TxResultMarshalled:    txResultMarshalled,
		BridgeCommitmentLeaf:  bcLeaf,
	}, nil
}

// fetchBridgeCommitmentLeaves takes an end exclusive range of heights and fetches its
// corresponding BridgeCommitmentLeafs.
func fetchBridgeCommitmentLeaves(
	ctx context.Context, clientCtx client.Context, start, end uint64,
) ([]types.BridgeCommitmentLeaf, error) {

	bridgeCommitmentLeaves := make([]types.BridgeCommitmentLeaf, 0, end-start)
	for height := start; height < end; height++ {

		int64Height := int64(height)
		commit, err := getCommit(ctx, clientCtx, &int64Height)
		if err != nil {
			return nil, fmt.Errorf("couldn't load block %d: %s", height, err.Error())
		}
		if commit == nil {
			return nil, fmt.Errorf("couldn't load commit %d", height)
		}

		bridgeCommitmentLeaves = append(bridgeCommitmentLeaves, types.BridgeCommitmentLeaf{
			Height:          height,
			LastResultsHash: commit.Header.LastResultsHash,
		})
	}

	return bridgeCommitmentLeaves, nil
}
