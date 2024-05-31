package commitments

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cometbft/cometbft/crypto/merkle"
	cmbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// NOTE: the server is currently queried as follows:
// http://localhost:1317/commitments/1/1000

type DataCommitmentsServer struct {
	clientCtx client.Context
}

func (dcs DataCommitmentsServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	values := strings.Split(req.URL.Path[1:], "/")
	v0, _ := strconv.Atoi(values[0])
	v1, _ := strconv.Atoi(values[1])
	//v2, _ := strconv.Atoi(values[2])
	//v3, _ := strconv.Atoi(values[3])

	//height, err := getBlockHeight(dcs.clientCtx)
	//if err != nil {
	//	panic(err)
	//}
	//height -= 1

	commitment, err := dcs.BridgeCommitment(dcs.clientCtx, uint64(v0), uint64(v1))
	if err != nil {
		panic(err)
	}

	//proof, err := dcs.BridgeCommitmentInclusionProof(dcs.clientCtx, int64(v2), int64(v3), uint64(v0), uint64(v1))
	//if err != nil {
	//	panic(err)
	//}

	//proofBz, err := json.Marshal(proof)
	//if err != nil {
	//	panic(err)
	//}

	result := map[string]interface{}{
		"commitment": commitment.BridgeCommitmentHash.String(),
		//"proof":      proofBz,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = fmt.Fprint(w, result)
	if err != nil {
		panic(err)
	}
}

func RegisterDataCommitmentsServer(clientCtx client.Context, rtr *mux.Router) {
	rtr.PathPrefix("/commitments").Handler(http.StripPrefix("/commitments", DataCommitmentsServer{clientCtx}))
}

func (dcs DataCommitmentsServer) BridgeCommitment(
	clientCtx client.Context, start, end uint64,
) (*ResultBridgeCommitment, error) {
	err := dcs.validateBridgeCommitmentRange(clientCtx, start, end)
	if err != nil {
		return nil, err
	}

	// Fetch data
	leaves, err := dcs.fetchBridgeCommitmentLeaves(clientCtx, start, end)
	if err != nil {
		return nil, err
	}
	// Encode data to match solidity side
	encodedLeaves, err := dcs.encodeBridgeCommitmentLeaves(leaves)
	if err != nil {
		return nil, err
	}
	// Generate merkle root
	root := merkle.HashFromByteSlices(encodedLeaves)

	return &ResultBridgeCommitment{
		BridgeCommitmentHash: root,
	}, nil
}

func (dcs DataCommitmentsServer) BridgeCommitmentInclusionProof(
	clientCtx client.Context, height, txIndex int64, start, end uint64,
) (*ResultBridgeCommitmentInclusionProof, error) {
	err := dcs.validateBridgeCommitmentInclusionProofRequest(clientCtx, uint64(height), start, end)
	if err != nil {
		return nil, err
	}

	// Fetch data
	leaves, err := dcs.fetchBridgeCommitmentLeaves(clientCtx, start, end)
	if err != nil {
		return nil, err
	}
	// Encode data to match solidity side
	encodedLeaves, err := dcs.encodeBridgeCommitmentLeaves(leaves)
	if err != nil {
		return nil, err
	}
	// Get proofs of the BridgeCommitment leaves
	_, proofs := merkle.ProofsFromByteSlices(encodedLeaves)

	// The Leaf of the BridgeCommitment the inclusion proof is for
	keyLeaf := leaves[height-int64(leaves[0].Height)]

	// Load the transactions that composed the LastResultsHash
	finalizeBlockResponse, err := getBlockResults(clientCtx, &height)
	if err != nil {
		return nil, err
	}

	if int(txIndex) >= len(finalizeBlockResponse.TxsResults) {
		return nil, fmt.Errorf("transaction index not found %d", txIndex)
	}
	deterministicTxResults := cmtypes.NewResults(finalizeBlockResponse.TxsResults)
	marshalledTxResult, err := deterministicTxResults[txIndex].Marshal()
	if err != nil {
		return nil, err
	}

	// Convert to HexBytes
	bcMerkleProof := *proofs[height-int64(leaves[0].Height)]
	bcAunts := make([]cmbytes.HexBytes, len(bcMerkleProof.Aunts))
	for i := range bcMerkleProof.Aunts {
		bcAunts[i] = bcMerkleProof.Aunts[i]
	}

	// Convert to HexBytes
	txMerkleProof := deterministicTxResults.ProveResult(int(txIndex))
	txAunts := make([]cmbytes.HexBytes, len(txMerkleProof.Aunts))
	for i := range txMerkleProof.Aunts {
		txAunts[i] = txMerkleProof.Aunts[i]
	}

	return &ResultBridgeCommitmentInclusionProof{
		BridgeCommitmentLeaf: keyLeaf,
		BridgeCommitmentMerkleProof: BinaryMerkleProof{
			Total:    bcMerkleProof.Total,
			Index:    bcMerkleProof.Index,
			LeafHash: bcMerkleProof.LeafHash,
			Aunts:    bcAunts,
		},
		ExecTxResult: marshalledTxResult,
		TxResultMerkleProof: BinaryMerkleProof{
			Total:    txMerkleProof.Total,
			Index:    txMerkleProof.Index,
			LeafHash: txMerkleProof.LeafHash,
			Aunts:    txAunts,
		},
	}, nil
}

func (dcs DataCommitmentsServer) fetchBridgeCommitmentLeaves(
	clientCtx client.Context, start, end uint64,
) ([]BridgeCommitmentLeaf, error) {

	bridgeCommitmentLeaves := make([]BridgeCommitmentLeaf, 0, end-start)
	for height := start; height < end; height++ {

		int64Height := int64(height)
		block, err := getBlock(clientCtx, &int64Height)
		if block == nil || err != nil {
			return nil, fmt.Errorf("couldn't load block %d", height)
		}

		// Load the next block to get the LastResultsHash since the hash is computed in the next block
		int64HeightPlusOne := int64(height + 1)
		nextBlock, err := getBlock(clientCtx, &int64HeightPlusOne)
		if block == nil || err != nil {
			return nil, fmt.Errorf("couldn't load block %d", height+1)
		}

		bridgeCommitmentLeaves = append(bridgeCommitmentLeaves, BridgeCommitmentLeaf{
			Height:      height,
			DataHash:    block.Block.Header.DataHash,
			ResultsHash: nextBlock.Block.Header.LastResultsHash,
		})
	}

	return bridgeCommitmentLeaves, nil
}

func (dcs DataCommitmentsServer) encodeBridgeCommitmentLeaves(leaves []BridgeCommitmentLeaf) ([][]byte, error) {

	encodedLeaves := make([][]byte, 0, len(leaves))
	for _, leaf := range leaves {

		// Pad to match `abi.encode` on Ethereum
		paddedHeight, err := To32PaddedHexBytes(leaf.Height)
		if err != nil {
			return nil, err
		}
		encodedHeightAndDataHash := append(paddedHeight, leaf.DataHash...)
		encodedLeaf := append(encodedHeightAndDataHash, leaf.ResultsHash...)

		encodedLeaves = append(encodedLeaves, encodedLeaf)
	}

	return encodedLeaves, nil
}

// To32PaddedHexBytes takes a number and returns its hex representation padded to 32 bytes.
// Used to mimic the result of `abi.encode(number)` in Ethereum.
func To32PaddedHexBytes(number uint64) ([]byte, error) {
	hexRepresentation := strconv.FormatUint(number, 16)
	// Make sure hex representation has even length.
	// The `strconv.FormatUint` can return odd length hex encodings.
	// For example, `strconv.FormatUint(10, 16)` returns `a`.
	// Thus, we need to pad it.
	if len(hexRepresentation)%2 == 1 {
		hexRepresentation = "0" + hexRepresentation
	}
	hexBytes, hexErr := hex.DecodeString(hexRepresentation)
	if hexErr != nil {
		return nil, hexErr
	}
	paddedBytes, padErr := padBytes(hexBytes, 32)
	if padErr != nil {
		return nil, padErr
	}
	return paddedBytes, nil
}

// padBytes Pad bytes to given length
func padBytes(byt []byte, length int) ([]byte, error) {
	l := len(byt)
	if l > length {
		return nil, fmt.Errorf(
			"cannot pad bytes because length of bytes array: %d is greater than given length: %d",
			l,
			length,
		)
	}
	if l == length {
		return byt, nil
	}
	tmp := make([]byte, length)
	copy(tmp[length-l:], byt)
	return tmp, nil
}
