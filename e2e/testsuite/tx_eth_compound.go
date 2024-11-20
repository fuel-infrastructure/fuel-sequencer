package testsuite

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
)

// DepositTokenToSequencer generates a deposit to an account owned by the sender by depositing V2 tokens, which results
// in a deposit with no lockup. The returned transaction receipt is the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DepositTokenToSequencer(amount *big.Int) *ethereumtypes.Receipt {

	// ...approve V2 tokens for use by sequencer interface contract.
	approveData := PackApproveToken(SequencerInterfaceContractAddress, amount)
	_, err := s.SendEthTransactionToTokenContract(approveData)
	s.Require().NoError(err)

	// ...deposit.
	depositData := PackDeposit(amount)
	receipt, err := s.SendEthTransactionToSequencerInterfaceContract(depositData)
	s.Require().NoError(err)

	return receipt
}

// DepositAndDelegateTokenToSequencer generates a depositAndDelegate to an account owned by the sender by depositing
// V2 tokens, which results in a deposit with no lockup. This is followed by an authorize which delegates the tokens.
// The returned transaction receipt is the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DepositAndDelegateTokenToSequencer(
	amount *big.Int, validator common.Address,
) *ethereumtypes.Receipt {

	// ...approve V2 tokens for use by sequencer interface contract.
	approveData := PackApproveToken(SequencerInterfaceContractAddress, amount)
	_, err := s.SendEthTransactionToTokenContract(approveData)
	s.Require().NoError(err)

	// ...deposit.
	depositData := PackDepositAndDelegate(amount, validator)
	receipt, err := s.SendEthTransactionToSequencerInterfaceContract(depositData)
	s.Require().NoError(err)

	return receipt
}

// DepositTokenToSequencerFromMigrationNoDelegation calls DepositTokenToSequencerFromMigration with a null Ethereum
// address as the validator to delegate to, meaning that there will be no delegation.
func (s *E2ETestSuite) DepositTokenToSequencerFromMigrationNoDelegation(
	amount *big.Int, vestingPeriod time.Duration,
) *ethereumtypes.Receipt {
	validator := common.HexToAddress(keeper.NullEthereumAddress)
	return s.DepositTokenToSequencerFromMigration(amount, validator, vestingPeriod)
}

// DepositTokenToSequencerFromMigration generates a deposit to an account owned by the sender by migrating V1 tokens to
// V2 tokens, which results in a deposit with lockup. The returned receipt is the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DepositTokenToSequencerFromMigration(
	amount *big.Int, validator common.Address, vestingPeriod time.Duration,
) *ethereumtypes.Receipt {

	// Amount needs to be upscaled by 1e9 to counteract the DECIMALS_DOWNSCALING_FACTOR of 1e9 applied by the migrator.
	amountToMigrate := new(big.Int).Mul(amount, MigrateAmountUpscalingFactor)

	// ...approve V1 tokens for use by token migrator.
	approveData := PackApproveMigratedToken(TokenMigratorContractAddress, amountToMigrate)
	_, err := s.SendEthTransactionToMigratedTokenContract(approveData)
	s.Require().NoError(err)

	// ...mint V2 tokens to token migrator.
	migratorMintData := PackMintToken(TokenMigratorContractAddress, amountToMigrate)
	_, err = s.SendEthTransactionToTokenContract(migratorMintData)
	s.Require().NoError(err)

	// ...migrate V1 tokens to V2 tokens.
	depositData := PackMigrate(amountToMigrate, validator, new(big.Int).SetInt64(int64(vestingPeriod.Seconds())))
	receipt, err := s.SendEthTransactionToTokenMigratorContract(depositData)
	s.Require().NoError(err)

	return receipt
}
