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

func (s *E2ETestSuite) GetFuelStreamXAddress() common.Address {
	return FuelStreamXContractAddress
}

func (s *E2ETestSuite) GetSequencerProxyAddress() common.Address {
	return SequencerProxyContractAddress
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
		return nil, fmt.Errorf("transaction failed - check Ethereum node logs; tx:%X; receipt:%s", txJson, receiptJson)
	}

	return receipt, nil
}
