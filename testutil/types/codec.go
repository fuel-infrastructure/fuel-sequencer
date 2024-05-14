package types

import (
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
)

var (
	encodingConfig testutil.TestEncodingConfig
	cdc            codec.Codec
	TestCdc        codec.Codec // an exported alias of cdc
)

func init() {
	// TODO: is this the correct way?
	modules := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distribution.AppModuleBasic{},
		consensus.AppModuleBasic{},
		slashing.AppModuleBasic{},
		mint.AppModuleBasic{},
		gov.AppModuleBasic{},
		crisis.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		bridge.AppModuleBasic{},
		sequencing.AppModuleBasic{},
	}
	encodingConfig = testutil.MakeTestEncodingConfig(modules...)

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

	cdc = encodingConfig.Codec
	TestCdc = cdc
}
