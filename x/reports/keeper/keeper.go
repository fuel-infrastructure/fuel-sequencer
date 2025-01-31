package keeper

import (
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

type (
	Keeper struct {
		appmodule.Environment

		cdc    codec.BinaryCodec
		logger log.Logger

		// Keepers
		stakingKeeper types.StakingKeeper

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	logger log.Logger,
	stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		Environment:   env,
		cdc:           cdc,
		authority:     authority,
		logger:        logger,
		stakingKeeper: stakingKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
