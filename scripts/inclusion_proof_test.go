package scripts_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/ethereum/go-ethereum/common"
	commitmentstypes "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BridgeCommitmentLeafForEthereum struct {
	Height      *big.Int
	ResultsHash common.Hash
}

type BinaryMerkleProofForEthereum struct {
	SideNodes []common.Hash
	Key       *big.Int
	NumLeaves *big.Int
}

func AuntsToHashes(proof merkle.Proof) (hashes []common.Hash) {
	hashes = make([]common.Hash, len(proof.Aunts))
	for i, aunt := range proof.Aunts {
		hashes[i] = common.BytesToHash(aunt)
	}
	return
}

func TestBridgeCommitmentInclusionProof(t *testing.T) {

	addr := "localhost:9090"
	height := int64(5)
	txIndex := int64(0)
	start := uint64(1)
	end := uint64(10)
	proofNonce := 0

	ctx := context.Background()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	client := commitmentstypes.NewQueryClient(conn)

	req := &commitmentstypes.QueryBridgeCommitmentInclusionProofRequest{
		Height:  height,
		TxIndex: txIndex,
		Start:   start,
		End:     end,
	}

	bridgeCommitmentInclusionProof, err := client.BridgeCommitmentInclusionProof(ctx, req)
	if err != nil {
		panic(err)
	}

	// Construct BridgeCommitment leaf proof from the inclusion proof data.

	bridgeCommitmentMerkleProof := bridgeCommitmentInclusionProof.BridgeCommitmentProof
	bridgeCommitmentLeafProof := BinaryMerkleProofForEthereum{
		SideNodes: AuntsToHashes(*bridgeCommitmentMerkleProof.ToMerkleProof()),
		Key:       big.NewInt(bridgeCommitmentMerkleProof.Index),
		NumLeaves: big.NewInt(bridgeCommitmentMerkleProof.Total),
	}

	// Construct tx result proof from the inclusion proof data.

	lastResultsMerkleProof := bridgeCommitmentInclusionProof.LastResultsProof
	txResultProof := BinaryMerkleProofForEthereum{
		SideNodes: AuntsToHashes(*lastResultsMerkleProof.ToMerkleProof()),
		Key:       big.NewInt(lastResultsMerkleProof.Index),
		NumLeaves: big.NewInt(lastResultsMerkleProof.Total),
	}

	// Construct BridgeCommitmentLeaf from the inclusion proof data.

	bridgeCommitmentLeaf := BridgeCommitmentLeafForEthereum{
		Height:      big.NewInt(int64(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.Height)),
		ResultsHash: common.BytesToHash(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash),
	}

	// Check that the proof is able to verify the marshalled tx result.

	lastResultsHash := bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash
	txResultMarshalled := bridgeCommitmentInclusionProof.TxResultMarshalled
	err = lastResultsMerkleProof.ToMerkleProof().Verify(lastResultsHash, txResultMarshalled)
	if err != nil {
		panic(err)
	}

	fmt.Printf("proofNonce: %d\n", proofNonce)
	fmt.Printf("bridgeCommitmentLeaf: %s\n", bridgeCommitmentLeaf)
	fmt.Printf("bridgeCommitmentLeafProof: %s\n", bridgeCommitmentLeafProof)
	fmt.Printf("txResultMarshalled: %s\n", txResultMarshalled)
	fmt.Printf("txResultProof: %s\n", txResultProof)
}
