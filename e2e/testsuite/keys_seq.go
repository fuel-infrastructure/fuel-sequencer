package testsuite

import (
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// SequencerKey holds a Sequencer account/validator address pair, as well as hex versions of each.
type SequencerKey struct {
	cryptotypes.PrivKey
	Address       sdk.AccAddress
	ValAddress    sdk.ValAddress
	AddressHex    string
	AddressSeq    string // Bech32
	ValAddressHex string
	ValAddressSeq string // Bech32
}

func mustNewSequencerKeyFromMnemonic(mnemonic string) *SequencerKey {
	key, err := newSequencerKeyFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}
	return key
}

func newSequencerKeyFromMnemonic(mnemonic string) (*SequencerKey, error) {
	kb := keyring.NewInMemory(cdc)

	name := "name"
	passphrase := ""

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	_, err = kb.NewAccount(name, mnemonic, passphrase, sdk.FullFundraiserPath, algo)
	if err != nil {
		return nil, err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return nil, err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return nil, err
	}

	address := sdk.AccAddress(privKey.PubKey().Address().Bytes())
	valAddress := sdk.ValAddress(privKey.PubKey().Address().Bytes())

	return &SequencerKey{
		PrivKey:       privKey,
		Address:       address,
		ValAddress:    valAddress,
		AddressHex:    common.Bytes2Hex(address.Bytes()),
		AddressSeq:    address.String(),
		ValAddressHex: common.Bytes2Hex(valAddress.Bytes()),
		ValAddressSeq: valAddress.String(),
	}, nil
}
