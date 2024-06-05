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
	return s.SendEthTransactionFrom(ETH_KEYS[0].PrivateKey, common.HexToAddress(TOKEN_CONTRACT), data)
}

func (s *E2ETestSuite) SendEthTransactionToSequencerInterfaceContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(ETH_KEYS[0].PrivateKey, common.HexToAddress(SEQUENCER_INTERFACE_CONTRACT), data)
}

func (s *E2ETestSuite) SendEthTransactionToFuelStreamXContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransactionFrom(ETH_GUARDIAN.PrivateKey, common.HexToAddress(FUEL_STREAM_X_CONTRACT), data)
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
