package testsuite

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// EthereumKey holds a private/public key pair and the resultant address, as well as hex versions of each.
type EthereumKey struct {
	PublicKey     *ecdsa.PublicKey
	PrivateKey    *ecdsa.PrivateKey
	Address       common.Address
	PublicKeyHex  string
	PrivateKeyHex string
	AddressHex    string
	AddressSeq    string // Bech32
}

func mustNewEthereumKeyFromMnemonic(mnemonic string) *EthereumKey {
	key, err := newEthereumKeyFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}
	return key
}

func newEthereumKeyFromMnemonic(mnemonic string) (*EthereumKey, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	path, err := hdwallet.ParseDerivationPath(DerivationPath)
	if err != nil {
		return nil, err
	}

	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := wallet.PrivateKeyBytes(account)
	if err != nil {
		return nil, err
	}
	privateKeyECDSA, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := wallet.PublicKeyBytes(account)
	if err != nil {
		return nil, err
	}
	publicKeyECDSA, err := wallet.PublicKey(account)
	if err != nil {
		return nil, err
	}

	address := account.Address

	addressBz, err := addressCdc.StringToBytes(address.Hex())
	if err != nil {
		return nil, err
	}
	addressSeq, err := addressCdc.BytesToString(addressBz)
	if err != nil {
		return nil, err
	}

	return &EthereumKey{
		PrivateKey:    privateKeyECDSA,
		PublicKey:     publicKeyECDSA,
		Address:       address,
		PrivateKeyHex: hexutil.Encode(privateKeyBytes),
		PublicKeyHex:  hexutil.Encode(publicKeyBytes),
		AddressHex:    address.Hex(),
		AddressSeq:    addressSeq,
	}, nil
}

func mustNewEthereumKeyFromPrivateKey(privateKey string) *EthereumKey {
	key, err := newEthereumKeyFromPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	return key
}

func newEthereumKeyFromPrivateKey(privateKey string) (*EthereumKey, error) {

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKeyECDSA)

	publicKeyECDSA, ok := privateKeyECDSA.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to assert type: public key is not of type *ecdsa.PublicKey")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	addressBz, err := addressCdc.StringToBytes(address.Hex())
	if err != nil {
		return nil, err
	}
	addressSeq, err := addressCdc.BytesToString(addressBz)
	if err != nil {
		return nil, err
	}

	return &EthereumKey{
		PrivateKey:    privateKeyECDSA,
		PublicKey:     publicKeyECDSA,
		Address:       address,
		PrivateKeyHex: hexutil.Encode(privateKeyBytes),
		PublicKeyHex:  hexutil.Encode(publicKeyBytes),
		AddressHex:    address.Hex(),
		AddressSeq:    addressSeq,
	}, nil
}
