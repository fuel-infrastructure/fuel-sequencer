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

// GetEthPrivateKeyHex is expected to return 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
// unless the mnemonic has been changed.
func (s *E2ETestSuite) GetEthPrivateKeyHex() string {
	return s.Chain.validators[0].ethereumKey.privateKey
}

func (s *E2ETestSuite) GetEthPrivateKey() *ecdsa.PrivateKey {
	privateKey, err := crypto.HexToECDSA(s.GetEthPrivateKeyHex()[2:])
	s.Require().NoError(err)

	return privateKey
}

func (s *E2ETestSuite) GetEthPublicKey() *ecdsa.PublicKey {
	publicKey := s.GetEthPrivateKey().Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	s.Require().True(ok, "error casting public key to ECDSA")

	return publicKeyECDSA
}

func (s *E2ETestSuite) SendEthTransactionToFuelStreamXContract(data []byte) (*ethereumtypes.Receipt, error) {
	return s.SendEthTransaction(common.HexToAddress(FUEL_STREAM_X_CONTRACT), data)
}

func (s *E2ETestSuite) SendEthTransaction(toAddress common.Address, data []byte) (*ethereumtypes.Receipt, error) {

	privateKey := s.GetEthPrivateKey()
	publicKey := s.GetEthPublicKey()

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

	s.Logger().Info(fmt.Sprintf("Submitting transaction to FuelStreamX contract: %s", signedTx.Hash().Hex()))
	err = s.Chain.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	// Ensure transaction went through by waiting 1 block.
	err = s.WaitForEthereumBlocks(s.Ctx(), 1, time.Second*10)
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
