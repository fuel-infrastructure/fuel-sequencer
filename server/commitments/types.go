package commitments

import "github.com/cometbft/cometbft/libs/bytes"

type BridgeCommitmentLeaf struct {
	Height      uint64         `json:"height"`
	DataHash    bytes.HexBytes `json:"data_hash"`
	ResultsHash bytes.HexBytes `json:"results_hash"`
}

type ResultBridgeCommitment struct {
	BridgeCommitmentHash bytes.HexBytes `json:"bridge_commitment_hash"`
}

type ResultBridgeCommitmentInclusionProof struct {

	// ------ Verify that a block was in BridgeCommitment

	// BridgeCommitmentLeaf is a leaf node in the BridgeCommitment merkle.
	BridgeCommitmentLeaf BridgeCommitmentLeaf `json:"bridge_commitment_leaf"`

	// BridgeCommitmentMerkleProof is the merkle proof to proof a BridgeCommitmentLeaf is in the BridgeCommitment
	// merkle tree.
	BridgeCommitmentMerkleProof BinaryMerkleProof `json:"bridge_commitment_merkle_proof"`

	// ------ Verify that a transaction was in TxResult

	// ExecTxResult is the deterministic response of the queried transaction.
	ExecTxResult bytes.HexBytes `json:"exec_tx_result"`
	// TxResultMerkleProof is the merkle proof to proof the result of a transaction if in the merkle tree.
	TxResultMerkleProof BinaryMerkleProof `json:"tx_result_merkle_proof"`
}

type BinaryMerkleProof struct {
	Total    int64            `json:"total"`     // Total number of items.
	Index    int64            `json:"index"`     // Index of item to prove.
	LeafHash bytes.HexBytes   `json:"leaf_hash"` // Hash of item value.
	Aunts    []bytes.HexBytes `json:"aunts"`     // Hashes from leaf's sibling to a root's child.
}
