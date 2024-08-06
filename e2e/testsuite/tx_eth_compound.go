package testsuite

import (
	"math/big"

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

// DepositTokenToSequencerFromMigration generates a deposit to an account owned by the sender by migrating V1 tokens to
// V2 tokens, which results in a deposit with lockup. The returned receipt is the one from the last transaction.
//
// Note: by default the sender is s.EthKeys[0]
func (s *E2ETestSuite) DepositTokenToSequencerFromMigration(amount *big.Int) *ethereumtypes.Receipt {

	// ...approve V1 tokens for use by token migrator.
	approveData := PackApproveMigratedToken(TokenMigratorContractAddress, amount)
	_, err := s.SendEthTransactionToMigratedTokenContract(approveData)
	s.Require().NoError(err)

	// ...mint V2 tokens to token migrator.
	migratorMintData := PackMintToken(TokenMigratorContractAddress, amount)
	_, err = s.SendEthTransactionToTokenContract(migratorMintData)
	s.Require().NoError(err)

	// ...migrate V1 tokens to V2 tokens.
	depositData := PackMigrate(amount, common.HexToAddress(keeper.NullEthereumAddress))
	receipt, err := s.SendEthTransactionToTokenMigratorContract(depositData)
	s.Require().NoError(err)

	return receipt
}
