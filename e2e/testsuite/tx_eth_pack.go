package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func PackBalanceOfERC20Token(tokenContractAbi string, address common.Address) []byte {
	return packCall(
		tokenContractAbi,
		BalanceOfQueryName,
		[]interface{}{
			address,
		},
	)
}

func PackApproveERC20Token(abi string, address common.Address, amount *big.Int) []byte {
	return packCall(
		abi,
		ApproveFunctionName,
		[]interface{}{
			address,
			amount,
		},
	)
}

func PackApproveToken(address common.Address, amount *big.Int) []byte {
	return PackApproveERC20Token(TokenContractABI, address, amount)
}

func PackApproveMigratedToken(address common.Address, amount *big.Int) []byte {
	return PackApproveERC20Token(MigratedTokenContractABI, address, amount)
}

func PackMintERC20Token(abi string, address common.Address, amount *big.Int) []byte {
	return packCall(
		abi,
		MintFunctionName,
		[]interface{}{
			address,
			amount,
		},
	)
}

func PackMintToken(address common.Address, amount *big.Int) []byte {
	return PackMintERC20Token(TokenContractABI, address, amount)
}

func PackMintMigratedToken(address common.Address, amount *big.Int) []byte {
	return PackMintERC20Token(MigratedTokenContractABI, address, amount)
}

func PackTransferAndCall(amount *big.Int) []byte {
	return packCall(
		TokenContractABI,
		TransferAndCallFunctionName,
		[]interface{}{
			SequencerInterfaceContractAddress,
			amount,
		},
	)
}

func PackMigrate(amount *big.Int, validator common.Address) []byte {
	return packCall(
		TokenMigratorContractABI,
		MigrateFunctionName,
		[]interface{}{
			amount,
			validator,
		},
	)
}

func PackAuthorize(data []byte) []byte {
	return PackBatchAuthorize([][]byte{data})
}

func PackBatchAuthorize(data [][]byte) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		BatchAuthorizeFunctionName,
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
		FuelStreamXContractABI,
		ProcessSequencerWithdrawalMessageFunctionName,
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
		FuelStreamXContractABI,
		UpdateGenesisStateFunctionName,
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
		FuelStreamXContractABI,
		UpdateCommitHeaderRangeFunctionName,
		[]interface{}{
			targetBlock,
			targetHeader,
			bridgeCommitment,
		},
	)
}
