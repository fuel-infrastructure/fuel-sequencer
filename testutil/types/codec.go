package types

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/bank"
	"cosmossdk.io/x/consensus"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/slashing"
	"cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"
	txdecode "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	appcodec "github.com/fuel-infrastructure/fuel-sequencer/app/codec"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
)

var (
	encodingConfig          testutil.TestEncodingConfig
	TestCdc                 codec.Codec
	TestAddressCdc          signing.AddressCodec
	TestValidatorAddressCdc address.ValidatorAddressCodec
	TestDecoder             *txdecode.Decoder
	TestTxDecoder           func([]byte) (sdk.Tx, error)
)

func init() {
	TestAddressCdc = appcodec.NewFuelSequencerAddressCodec(
		sdkAddressCodec.NewBech32Codec(app.AccountAddressPrefix),
	)
	TestValidatorAddressCdc = appcodec.NewFuelSequencerAddressCodec(
		sdkAddressCodec.NewBech32Codec(app.AccountAddressPrefix + "valoper"),
	)

	modules := []module.AppModule{
		auth.AppModule{},
		accounts.AppModule{},
		bank.AppModule{},
		staking.AppModule{},
		distribution.AppModule{},
		consensus.AppModule{},
		slashing.AppModule{},
		mint.AppModule{},
		gov.AppModule{},
		upgrade.AppModule{},
		evidence.AppModule{},
		bridge.AppModule{},
		sequencing.AppModule{},
	}
	encodingConfig = testutil.MakeTestEncodingConfig(codectestutil.CodecOptions{
		AccAddressPrefix: app.AccountAddressPrefix,
		ValAddressPrefix: app.AccountAddressPrefix + "valoper",
		AddressCodec:     TestAddressCdc,
		ValidatorCodec:   TestValidatorAddressCdc,
	}, modules...)

	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
		&stakingtypes.MsgBeginRedelegate{},
	)
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*cryptotypes.PubKey)(nil),
		&secp256k1.PubKey{},
		&ed25519.PubKey{},
	)
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.AccountI)(nil),
		&bridgetypes.EthOwnedBaseAccount{},
		&bridgetypes.EthOwnedContinuousVestingAccount{},
		&vestingtypes.DelayedVestingAccount{},
		&vestingtypes.ContinuousVestingAccount{},
	)

	TestCdc = encodingConfig.Codec
	decoder, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: TestCdc.InterfaceRegistry().SigningContext(),
		ProtoCodec:     TestCdc,
	})
	if err != nil {
		panic(err)
	}
	TestDecoder = decoder

	TestTxDecoder = authtx.DefaultTxDecoder(TestAddressCdc, TestCdc, TestDecoder)
}
