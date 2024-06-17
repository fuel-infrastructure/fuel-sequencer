package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func PackMint(address common.Address, amount *big.Int) []byte {
	return packCall(
		sidecartypes.TokenContractABI,
		sidecartypes.MintFunctionName,
		[]interface{}{
			address,
			amount,
		},
	)
}

func PackTransferAndCall(amount *big.Int) []byte {
	return packCall(
		sidecartypes.TokenContractABI,
		sidecartypes.TransferAndCallFunctionName,
		[]interface{}{
			common.HexToAddress(SEQUENCER_INTERFACE_CONTRACT),
			amount,
		},
	)
}

func PackAuthorize(data []byte) []byte {
	return PackBatchAuthorize([][]byte{data})
}

func PackBatchAuthorize(data [][]byte) []byte {
	return packCall(
		sidecartypes.SequencerInterfaceContractABI,
		sidecartypes.BatchAuthorizeFunctionName,
		[]interface{}{
			data,
		},
	)
}

func PackProcessSequencerWithdrawalMessage(
	proofNonce *big.Int,
	bridgeCommitmentLeaf BridgeCommitmentLeafForEthereum,
	bridgeCommitmentLeafProof BinaryMerkleProofForEthereum,
	txResultMarshalled []byte,
	txResultProof BinaryMerkleProofForEthereum,
) []byte {
	return packCall(
		sidecartypes.FuelStreamXContractABI,
		sidecartypes.ProcessSequencerWithdrawalMessageFunctionName,
		[]interface{}{
			proofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		},
	)
}

func PackUpdateGenesisStateMessage(
	height uint32, trustedHeader common.Hash,
) []byte {
	return packCall(
		sidecartypes.FuelStreamXContractABI,
		sidecartypes.UpdateGenesisStateFunctionName,
		[]interface{}{
			height,
			trustedHeader,
		},
	)
}

func PackUpdateCommitHeaderRangeMessage(
	targetBlock uint64,
	targetHeader common.Hash,
	bridgeCommitment common.Hash,
) []byte {
	return packCall(
		sidecartypes.FuelStreamXContractABI,
		sidecartypes.UpdateCommitHeaderRangeFunctionName,
		[]interface{}{
			targetBlock,
			targetHeader,
			bridgeCommitment,
		},
	)
}
