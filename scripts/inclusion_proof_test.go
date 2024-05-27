package scripts_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/merkle"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/ethereum/go-ethereum/common"
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

	addr := "https://rpc-seq.simplystaking.xyz"
	height := int64(89290)
	txIndex := int64(1)
	start := uint64(89142)
	end := uint64(89364)
	proofNonce := 98

	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	bridgeCommitmentInclusionProof, err := rpcClient.BridgeCommitmentInclusionProof(ctx, height, txIndex, start, end)
	if err != nil {
		panic(err)
	}

	// Construct BridgeCommitment leaf proof from the inclusion proof data.

	bridgeCommitmentMerkleProof := bridgeCommitmentInclusionProof.BridgeCommitmentMerkleProof
	bridgeCommitmentLeafProof := BinaryMerkleProofForEthereum{
		SideNodes: AuntsToHashes(*bridgeCommitmentMerkleProof.ToMerkleProof()),
		Key:       big.NewInt(bridgeCommitmentMerkleProof.Index),
		NumLeaves: big.NewInt(bridgeCommitmentMerkleProof.Total),
	}

	// Construct tx result proof from the inclusion proof data.

	lastResultsMerkleProof := bridgeCommitmentInclusionProof.LastResultsMerkleProof
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

	fmt.Println(fmt.Sprintf("proofNonce: %d", proofNonce))
	fmt.Println(fmt.Sprintf("bridgeCommitmentLeaf: %s", bridgeCommitmentLeaf))
	fmt.Println(fmt.Sprintf("bridgeCommitmentLeafProof: %s", bridgeCommitmentLeafProof))
	fmt.Println(fmt.Sprintf("txResultMarshalled: %s", txResultMarshalled))
	fmt.Println(fmt.Sprintf("txResultProof: %s", txResultProof))
}
