package testsuite

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"

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

func (s *E2ETestSuite) SendEthTransactionToProxyContract(data []byte) error {
	return s.SendEthTransaction(common.HexToAddress(FUEL_STREAM_X_CONTRACT), data)
}

func (s *E2ETestSuite) SendEthTransaction(toAddress common.Address, data []byte) error {

	privateKey := s.GetEthPrivateKey()
	publicKey := s.GetEthPublicKey()

	fromAddress := crypto.PubkeyToAddress(*publicKey)
	nonce, err := s.Chain.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	value := big.NewInt(0)
	gasLimit := uint64(1000000)
	gasPrice, err := s.Chain.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return err
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
		return err
	}

	signedTx, err := ethereumtypes.SignTx(tx, ethereumtypes.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}

	err = s.Chain.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	return nil
}
