package testsuite

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
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

func (s *E2ETestSuite) DepositForToSequencer(
	amount *big.Int, recipient common.Address,
) *ethereumtypes.Receipt {

	// ...approve V2 tokens for use by sequencer interface contract.
	approveData := PackApproveToken(SequencerInterfaceContractAddress, amount)
	_, err := s.SendEthTransactionToTokenContract(approveData)
	s.Require().NoError(err)

	// ...deposit.
	depositData := PackDepositFor(amount, recipient)
	receipt, err := s.SendEthTransactionToSequencerInterfaceContract(depositData)
	s.Require().NoError(err)

	return receipt
}

// DelegateTokenToSequencer generates a delegation to a validator on the Sequencer. The returned transaction receipt is
// the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DelegateTokenToSequencer(amount *big.Int, validator common.Address) *ethereumtypes.Receipt {
	// ...delegate.
	delegateData := PackDelegate(amount, validator)
	receipt, err := s.SendEthTransactionToSequencerInterfaceContract(delegateData)
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

// DepositTokenToSequencerFromMigrationNoDelegation generates a deposit to an account owned by the sender by migrating
// V1 tokens to V2 tokens, which results in a deposit with lockup. The returned transaction receipt is the one from the
// last transaction.
//
// Notes:
// 1. by default the sender is s.EthKeys[0]
// 2. by default, the migrate function does not generate a delegation
func (s *E2ETestSuite) DepositTokenToSequencerFromMigrationNoDelegation(
	amount *big.Int, vestingPeriod time.Duration,
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
	depositData := PackMigrate(amountToMigrate, new(big.Int).SetInt64(int64(vestingPeriod.Seconds())))
	receipt, err := s.SendEthTransactionToTokenMigratorContract(depositData)
	s.Require().NoError(err)

	return receipt
}

// DepositTokenToSequencerFromMigrationAndDelegate generates a deposit to an account owned by the sender by first
// migrating V1 tokens to V2 tokens, which results in a deposit with lockup, and then initiates a delegation. The
// returned transaction receipt is the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DepositTokenToSequencerFromMigrationAndDelegate(
	amount *big.Int, validator common.Address, vestingPeriod time.Duration,
) *ethereumtypes.Receipt {
	// ... migrate tokens
	s.DepositTokenToSequencerFromMigrationNoDelegation(amount, vestingPeriod)

	// ...delegate tokens on Sequencer. Amount needs to take into consideration the migration ratio.
	amountToDelegate := new(big.Int).Mul(amount, MigrationRatio)
	receipt := s.DelegateTokenToSequencer(amountToDelegate, validator)

	return receipt
}
