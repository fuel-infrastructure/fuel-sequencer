package service

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/cometbft/cometbft/crypto/merkle"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

type queryServer struct {
	clientCtx         client.Context
	interfaceRegistry codectypes.InterfaceRegistry
}

func NewQueryServer(
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
) types.QueryServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
	}
}

func (q queryServer) BridgeCommitment(
	_ context.Context, req *types.QueryBridgeCommitmentRequest,
) (*types.QueryBridgeCommitmentResponse, error) {

	err := validateBridgeCommitmentRange(q.clientCtx, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	// Fetch data
	leaves, err := fetchBridgeCommitmentLeaves(q.clientCtx, req.Start, req.End)
	if err != nil {
		return nil, err
	}
	// Encode data to match solidity side
	encodedLeaves, err := encodeBridgeCommitment(leaves)
	if err != nil {
		return nil, err
	}
	// Generate merkle root
	root := merkle.HashFromByteSlices(encodedLeaves)

	return &types.QueryBridgeCommitmentResponse{
		BridgeCommitmentHash: hex.EncodeToString(root),
	}, nil
}

func (q queryServer) BridgeCommitmentInclusionProof(
	_ context.Context, req *types.QueryBridgeCommitmentInclusionProofRequest,
) (*types.QueryBridgeCommitmentInclusionProofResponse, error) {
	err := validateBridgeCommitmentInclusionProofRequest(q.clientCtx, uint64(req.Height), req.Start, req.End)
	if err != nil {
		return nil, err
	}

	// Fetch data
	leaves, err := fetchBridgeCommitmentLeaves(q.clientCtx, req.Start, req.End)
	if err != nil {
		return nil, err
	}
	// Encode data to match solidity side
	encodedLeaves, err := encodeBridgeCommitment(leaves)
	if err != nil {
		return nil, err
	}
	// Get proofs of the BridgeCommitment leaves
	_, proofs := merkle.ProofsFromByteSlices(encodedLeaves)

	// The Leaf of the BridgeCommitment the inclusion proof is for
	keyLeaf := leaves[req.Height-int64(leaves[0].Height)]

	// Load the transactions that composed the LastResultsHash
	txResultHeight := req.Height - 1
	finalizeBlockResponse, err := getBlockResults(q.clientCtx, &txResultHeight)
	if err != nil {
		return nil, err
	}

	if int(req.TxIndex) >= len(finalizeBlockResponse.TxsResults) {
		return nil, fmt.Errorf("transaction index not found %d", req.TxIndex)
	}
	deterministicTxResults := cmtypes.NewResults(finalizeBlockResponse.TxsResults)
	marshalledTxResult, err := deterministicTxResults[req.TxIndex].Marshal()
	if err != nil {
		return nil, err
	}

	// Convert to strings
	bcMerkleProof := *proofs[req.Height-int64(leaves[0].Height)]
	bcAunts := make([]string, len(bcMerkleProof.Aunts))
	for i := range bcMerkleProof.Aunts {
		bcAunts[i] = hex.EncodeToString(bcMerkleProof.Aunts[i])
	}

	// Convert to strings
	txMerkleProof := deterministicTxResults.ProveResult(int(req.TxIndex))
	txAunts := make([]string, len(txMerkleProof.Aunts))
	for i := range txMerkleProof.Aunts {
		txAunts[i] = hex.EncodeToString(txMerkleProof.Aunts[i])
	}

	return &types.QueryBridgeCommitmentInclusionProofResponse{
		BridgeCommitmentLeaf: &types.BridgeCommitmentLeaf{
			Height:          keyLeaf.Height,
			LastResultsHash: hex.EncodeToString(keyLeaf.LastResultsHash),
		},
		BridgeCommitmentProof: &types.BinaryMerkleProof{
			Total:    bcMerkleProof.Total,
			Index:    bcMerkleProof.Index,
			LeafHash: hex.EncodeToString(bcMerkleProof.LeafHash),
			Aunts:    bcAunts,
		},
		TxResultMarshalled: hex.EncodeToString(marshalledTxResult),
		LastResultsProof: &types.BinaryMerkleProof{
			Total:    txMerkleProof.Total,
			Index:    txMerkleProof.Index,
			LeafHash: hex.EncodeToString(txMerkleProof.LeafHash),
			Aunts:    txAunts,
		},
	}, nil
}

func fetchBridgeCommitmentLeaves(
	clientCtx client.Context, start, end uint64,
) ([]types.BridgeCommitmentLeafRaw, error) {

	bridgeCommitmentLeaves := make([]types.BridgeCommitmentLeafRaw, 0, end-start)
	for height := start; height < end; height++ {

		int64Height := int64(height)
		block, err := getBlock(clientCtx, &int64Height)
		if block == nil || err != nil {
			return nil, fmt.Errorf("couldn't load block %d", height)
		}

		bridgeCommitmentLeaves = append(bridgeCommitmentLeaves, types.BridgeCommitmentLeafRaw{
			Height:          height,
			LastResultsHash: block.Block.Header.LastResultsHash,
		})
	}

	return bridgeCommitmentLeaves, nil
}
