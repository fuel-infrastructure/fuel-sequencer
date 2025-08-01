package testsuite

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// packCall produces call data from a contract ABI, method, and args.
func packCall(abiString, method string, args []interface{}) []byte {
	encodedCall, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		panic(errorsmod.Wrap(err, "bad ABI definition in code"))
	}
	abiEncodedCall, err := encodedCall.Pack(method, args...)
	if err != nil {
		panic(errorsmod.Wrap(err, "error packing calling"))
	}
	return abiEncodedCall
}

func (s *E2ETestSuite) SendEthTransactionToTokenContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthKeys[0].PrivateKey, TokenContractAddress, data)
}

func (s *E2ETestSuite) SendEthTransactionToSequencerInterfaceContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthKeys[0].PrivateKey, SequencerInterfaceContractAddress, data)
}

// SendEthTransactionToFuelStreamXContractAsUser sends a transaction to FuelStreamX as the anvil-generated key at index
// 1, which does not have elevated privileges. The use of index 1 further distinguishes this function from:
// - SendEthTransactionToFuelStreamXContractAsDeployer which uses the anvil-generated key at index 0 (== s.EthKeys[0]).
// - SendEthTransactionToFuelStreamXContractAsGuardian which uses the anvil-generated key at index 19.
func (s *E2ETestSuite) SendEthTransactionToFuelStreamXContractAsUser(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthUser.PrivateKey, FuelStreamXContractAddress, data)
}

// SendEthTransactionToFuelStreamXContractAsDeployer sends a transaction to FuelStreamX as the contract deployer.
func (s *E2ETestSuite) SendEthTransactionToFuelStreamXContractAsDeployer(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthDeployer.PrivateKey, FuelStreamXContractAddress, data)
}

// SendEthTransactionToFuelStreamXContractAsGuardian sends a transaction to FuelStreamX as the contract guardian.
func (s *E2ETestSuite) SendEthTransactionToFuelStreamXContractAsGuardian(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthGuardian.PrivateKey, FuelStreamXContractAddress, data)
}

func (s *E2ETestSuite) SendEthTransactionToMigratedTokenContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthKeys[0].PrivateKey, MigratedTokenContractAddress, data)
}

func (s *E2ETestSuite) SendEthTransactionToTokenMigratorContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(s.EthKeys[0].PrivateKey, TokenMigratorContractAddress, data)
}

// SendEthTransactionFromToTokenContract sends a transaction to the Token contract from a specific private key
func (s *E2ETestSuite) SendEthTransactionFromToTokenContract(privateKey *ecdsa.PrivateKey, data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(privateKey, TokenContractAddress, data)
}

// SendEthTransactionFromToSequencerInterfaceContract sends a transaction to the SequencerInterface contract from a specific private key
func (s *E2ETestSuite) SendEthTransactionFromToSequencerInterfaceContract(privateKey *ecdsa.PrivateKey, data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(privateKey, SequencerInterfaceContractAddress, data)
}

// SendEthTransactionFromToMigratedTokenContract sends a transaction to the MigratedToken contract from a specific private key
func (s *E2ETestSuite) SendEthTransactionFromToMigratedTokenContract(privateKey *ecdsa.PrivateKey, data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(privateKey, MigratedTokenContractAddress, data)
}

// SendEthTransactionFromToTokenMigratorContract sends a transaction to the TokenMigrator contract from a specific private key
func (s *E2ETestSuite) SendEthTransactionFromToTokenMigratorContract(privateKey *ecdsa.PrivateKey, data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(privateKey, TokenMigratorContractAddress, data)
}

func (s *E2ETestSuite) SendEthTransactionFrom(privateKey *ecdsa.PrivateKey, toAddress common.Address, data []byte) (*ethereumtypes.Receipt, error) {

	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKey)
	nonce, err := s.Chain.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0)
	gasLimit := uint64(1000000)
	gasPrice, err := s.Chain.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	tx := ethereumtypes.NewTx(&ethereumtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	chainID, err := s.Chain.ethClient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := ethereumtypes.SignTx(tx, ethereumtypes.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	s.Logger().Info(fmt.Sprintf("Submitting transaction to contract: %s", signedTx.Hash().Hex()))
	err = s.Chain.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	// Wait for one block to ensure Ethereum transaction went through.
	err = s.WaitForEthereumBlocks(s.Ctx(), 1, time.Minute)
	s.Require().NoError(err)

	receipt, err := s.Chain.ethClient.TransactionReceipt(context.Background(), signedTx.Hash())
	if err != nil {
		return nil, err
	} else if receipt.Status != 1 {
		txJson, err := signedTx.MarshalJSON()
		s.Require().NoError(err)
		receiptJson, err := receipt.MarshalJSON()
		s.Require().NoError(err)
		return nil, fmt.Errorf("transaction failed - check Ethereum node logs; height:%d; tx:%X; receipt:%s", receipt.BlockNumber, txJson, receiptJson)
	}

	return receipt, nil
}

// SendEthTransactionFromWithValue sends an Ethereum transaction with ETH value from a specific private key
func (s *E2ETestSuite) SendEthTransactionFromWithValue(privateKey *ecdsa.PrivateKey, toAddress common.Address, data []byte, value *big.Int) (*ethereumtypes.Receipt, error) {
	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKey)
	nonce, err := s.Chain.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000) // Standard ETH transfer gas limit
	if len(data) > 0 {
		gasLimit = uint64(1000000) // Higher gas limit for contract calls
	}

	gasPrice, err := s.Chain.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	tx := ethereumtypes.NewTx(&ethereumtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	chainID, err := s.Chain.ethClient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := ethereumtypes.SignTx(tx, ethereumtypes.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	s.Logger().Info(fmt.Sprintf("Submitting ETH transfer: %s ETH to %s", value.String(), toAddress.Hex()))
	err = s.Chain.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	// Wait for one block to ensure Ethereum transaction went through
	err = s.WaitForEthereumBlocks(s.Ctx(), 1, time.Minute)
	if err != nil {
		return nil, err
	}

	receipt, err := s.Chain.ethClient.TransactionReceipt(context.Background(), signedTx.Hash())
	if err != nil {
		return nil, err
	} else if receipt.Status != 1 {
		return nil, fmt.Errorf("ETH transfer failed - tx hash: %s", signedTx.Hash().Hex())
	}

	return receipt, nil
}

// EnsureEthereumMinimumBalances ensures that all provided Ethereum keys have at least the minimum balance,
// redistributing ETH from accounts with excess to accounts with insufficient balance.
func (s *E2ETestSuite) EnsureEthereumMinimumBalances(
	minimumBalance *big.Int, keys ...*EthereumKey,
) error {
	s.T().Log("💰 Checking Ethereum balances...")

	// Track balances and categorize accounts
	type keyBalance struct {
		key     *EthereumKey
		balance *big.Int
	}

	sufficientKeys := make([]keyBalance, 0)
	insufficientKeys := make([]keyBalance, 0)
	totalBalance := big.NewInt(0)

	for i, key := range keys {
		s.T().Logf("   Checking EthKey[%d] (%s)", i, key.AddressHex)

		// Get ETH balance
		balance, err := s.QueryEthereumEthBalance(s.Ctx(), key.Address, nil)
		if err != nil {
			return fmt.Errorf("failed to query balance for key %d: %w", i, err)
		}
		s.T().Logf("   ✅ EthKey[%d] has %s ETH", i, balance.String())

		if balance.Cmp(minimumBalance) >= 0 {
			sufficientKeys = append(sufficientKeys, keyBalance{key, balance})
		} else {
			insufficientKeys = append(insufficientKeys, keyBalance{key, balance})
		}

		totalBalance.Add(totalBalance, balance)
	}

	s.T().Log("✅ Ethereum balances checked")

	// Check if total balance is sufficient
	requiredTotalBalance := new(big.Int).Mul(minimumBalance, big.NewInt(int64(len(keys))))
	if totalBalance.Cmp(requiredTotalBalance) < 0 {
		return fmt.Errorf("total balance (%s ETH) is less than required balance (%s ETH)",
			totalBalance.String(), requiredTotalBalance.String())
	}

	// If all accounts have sufficient balance, no redistribution needed
	if len(insufficientKeys) == 0 {
		s.T().Logf("✅ All accounts have a sufficient balance of %s ETH", minimumBalance.String())
		return nil
	}

	s.T().Logf("💸 Need to redistribute ETH: %d accounts need more, %d have excess",
		len(insufficientKeys), len(sufficientKeys))

	// Redistribute from accounts with excess to accounts with insufficient balance
	for _, insufficient := range insufficientKeys {
		needed := new(big.Int).Sub(minimumBalance, insufficient.balance)
		s.T().Logf("   Account %s needs %s ETH", insufficient.key.AddressHex, needed.String())

		// Find a sufficient account to transfer from
		for i, sufficient := range sufficientKeys {
			available := new(big.Int).Sub(sufficient.balance, minimumBalance)

			// If this account has enough to cover the need
			if available.Cmp(needed) >= 0 {
				s.T().Logf("   Transferring %s ETH from %s to %s",
					needed.String(), sufficient.key.AddressHex, insufficient.key.AddressHex)

				// Perform the ETH transfer
				err := s.sendEthTransfer(sufficient.key, insufficient.key.Address, needed)
				if err != nil {
					return fmt.Errorf("failed to transfer ETH from %s to %s: %w",
						sufficient.key.AddressHex, insufficient.key.AddressHex, err)
				}

				// Update the sufficient account's balance
				sufficientKeys[i].balance.Sub(sufficientKeys[i].balance, needed)
				break
			}
		}
	}

	s.T().Log("✅ ETH redistribution completed successfully")

	// Wait a bit for transactions to be mined
	s.T().Log("⏳ Waiting for ETH transfers to be processed...")
	time.Sleep(5 * time.Second)

	// Verify all accounts now have minimum balance
	for i, key := range keys {
		balance, err := s.QueryEthereumEthBalance(s.Ctx(), key.Address, nil)
		if err != nil {
			return fmt.Errorf("failed to verify balance for key %d: %w", i, err)
		}

		if balance.Cmp(minimumBalance) < 0 {
			return fmt.Errorf("account %s still has insufficient balance: %s ETH < %s ETH",
				key.AddressHex, balance.String(), minimumBalance.String())
		}

		s.T().Logf("   ✅ EthKey[%d] verified: %s ETH", i, balance.String())
	}

	return nil
}

// sendEthTransfer sends ETH from one account to another
func (s *E2ETestSuite) sendEthTransfer(fromKey *EthereumKey, toAddress common.Address, amount *big.Int) error {
	// Use empty data for simple ETH transfer - using testsuite method
	receipt, err := s.SendEthTransactionFromWithValue(fromKey.PrivateKey, toAddress, nil, amount)
	if err != nil {
		return err
	}

	// Wait for the transaction to be processed
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, receipt.BlockNumber.Uint64())

	return nil
}
