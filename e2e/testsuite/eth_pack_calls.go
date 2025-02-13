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

func PackHasRole(abi string, role common.Hash, address common.Address) []byte {
	return packCall(
		abi,
		HasRoleQueryName,
		[]interface{}{
			role,
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

func PackMigrate(amount *big.Int, vestingPeriod *big.Int) []byte {
	return packCall(
		TokenMigratorContractABI,
		MigrateFunctionName,
		[]interface{}{
			amount,
			vestingPeriod,
		},
	)
}

func PackDeposit(amount *big.Int) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		DepositFunctionName,
		[]interface{}{
			amount,
		},
	)
}

func PackDepositFor(amount *big.Int, recipient common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		DepositForFunctionName,
		[]interface{}{
			amount,
			recipient,
		},
	)
}

func PackDelegate(amount *big.Int, validator common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		DelegateFunctionName,
		[]interface{}{
			amount,
			validator,
		},
	)
}

func PackDepositAndDelegate(amount *big.Int, validator common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		DepositAndDelegateFunctionName,
		[]interface{}{
			amount,
			validator,
		},
	)
}

func PackRedelegate(amount *big.Int, srcValidator, dstValidator common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		RedelegateFunctionName,
		[]interface{}{
			amount,
			srcValidator,
			dstValidator,
		},
	)
}

func PackClaimRewards(validator common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		ClaimRewardsFunctionName,
		[]interface{}{
			validator,
		},
	)
}

func PackUnbond(amount *big.Int, validator common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		UnbondFunctionName,
		[]interface{}{
			amount,
			validator,
		},
	)
}

func PackWithdraw(amount *big.Int) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		WithdrawFunctionName,
		[]interface{}{
			amount,
		},
	)
}

func PackWithdrawTo(amount *big.Int, recipient common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		WithdrawToFunctionName,
		[]interface{}{
			amount,
			recipient,
		},
	)
}

func PackTransfer(recipient common.Address, amount *big.Int) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		TransferFunctionName,
		[]interface{}{
			recipient,
			amount,
		},
	)
}

func PackVote(proposalId uint64, option uint32, memory string) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		VoteFunctionName,
		[]interface{}{
			proposalId,
			option,
			memory,
		},
	)
}

func PackSetRewardRecipient(recipient common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		SetRewardRecipientFunctionName,
		[]interface{}{
			recipient,
		},
	)
}

func PackGrantClaimRewards(grantee common.Address, expiration uint32) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		GrantClaimRewardsFunctionName,
		[]interface{}{
			grantee,
			expiration,
		},
	)
}

func PackRevokeClaimRewards(grantee common.Address) []byte {
	return packCall(
		SequencerInterfaceContractABI,
		RevokeClaimRewardsFunctionName,
		[]interface{}{
			grantee,
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

func PackProcessSequencerSupplyDeltaMessage(
	proofNonce *big.Int,
	bridgeCommitmentLeaf BridgeCommitmentLeafForEthereum,
	bridgeCommitmentLeafProof BinaryMerkleProofForEthereum,
	txResultMarshalled []byte,
	txResultProof BinaryMerkleProofForEthereum,
) []byte {
	return packCall(
		FuelStreamXContractABI,
		ProcessSequencerSupplyDeltaMessageFunctionName,
		[]interface{}{
			proofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		},
	)
}
