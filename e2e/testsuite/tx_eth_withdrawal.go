package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const processSequencerWithdrawalMessageABIJSON = `
[
  {
    "type": "function",
    "name": "processSequencerWithdrawalMessage",
    "inputs": [
      {
        "name": "_proofNonce",
        "type": "uint256",
        "internalType": "uint256"
      },
      {
        "name": "bridgeCommitmentLeaf",
        "type": "tuple",
        "internalType": "struct FuelStreamX.BridgeCommitmentLeaf",
        "components": [
          {
            "name": "height",
            "type": "uint256",
            "internalType": "uint256"
          },
          {
            "name": "dataHash",
            "type": "bytes32",
            "internalType": "bytes32"
          },
          {
            "name": "resultsHash",
            "type": "bytes32",
            "internalType": "bytes32"
          }
        ]
      },
      {
        "name": "bridgeCommitmentLeafProof",
        "type": "tuple",
        "internalType": "struct BinaryMerkleProof",
        "components": [
          {
            "name": "sideNodes",
            "type": "bytes32[]",
            "internalType": "bytes32[]"
          },
          {
            "name": "key",
            "type": "uint256",
            "internalType": "uint256"
          },
          {
            "name": "numLeaves",
            "type": "uint256",
            "internalType": "uint256"
          }
        ]
      },
      {
        "name": "txResultMarshalled",
        "type": "bytes",
        "internalType": "bytes"
      },
      {
        "name": "txResultProof",
        "type": "tuple",
        "internalType": "struct BinaryMerkleProof",
        "components": [
          {
            "name": "sideNodes",
            "type": "bytes32[]",
            "internalType": "bytes32[]"
          },
          {
            "name": "key",
            "type": "uint256",
            "internalType": "uint256"
          },
          {
            "name": "numLeaves",
            "type": "uint256",
            "internalType": "uint256"
          }
        ]
      }
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  }
]
`

type BridgeCommitmentLeafForEthereum struct {
	Height      *big.Int
	DataHash    common.Hash
	ResultsHash common.Hash
}

type BinaryMerkleProofForEthereum struct {
	SideNodes []common.Hash
	Key       *big.Int
	NumLeaves *big.Int
}

func PackProcessSequencerWithdrawalMessage(
	proofNonce *big.Int,
	bridgeCommitmentLeaf BridgeCommitmentLeafForEthereum,
	bridgeCommitmentLeafProof BinaryMerkleProofForEthereum,
	txResultMarshalled []byte,
	txResultProof BinaryMerkleProofForEthereum,
) []byte {
	return packCall(processSequencerWithdrawalMessageABIJSON, "processSequencerWithdrawalMessage", []interface{}{
		proofNonce,
		bridgeCommitmentLeaf,
		bridgeCommitmentLeafProof,
		txResultMarshalled,
		txResultProof,
	})
}
