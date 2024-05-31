package keeper

import (
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// Keepers
		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		stakingKeeper types.StakingKeeper

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// blockedAddresses is the list of blocked addresses
		blockedAddresses map[string]bool

		// Msg server router
		router *baseapp.MsgServiceRouter
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	authority string,
	blockedAddresses map[string]bool,
	router *baseapp.MsgServiceRouter,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:              cdc,
		storeService:     storeService,
		authority:        authority,
		logger:           logger,
		bankKeeper:       bankKeeper,
		accountKeeper:    accountKeeper,
		stakingKeeper:    stakingKeeper,
		blockedAddresses: blockedAddresses,
		router:           router,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetAddressCodec returns the module's AccountKeeper.AddressCodec.
func (k Keeper) GetAddressCodec() address.Codec {
	return k.accountKeeper.AddressCodec()
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
