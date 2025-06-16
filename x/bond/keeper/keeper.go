package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// keepers
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		bridgeKeeper  types.BridgeKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	bridgeKeeper types.BridgeKeeper,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:           cdc,
		storeService:  storeService,
		authority:     authority,
		logger:        logger,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		bridgeKeeper:  bridgeKeeper,
	}
}

// GetAccountKeeper returns the account keeper
func (k Keeper) GetAccountKeeper() types.AccountKeeper {
	return k.accountKeeper
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetAddressCodec returns the module's AccountKeeper.AddressCodec.
func (k Keeper) GetAddressCodec() address.Codec {
	return k.accountKeeper.AddressCodec()
}

func (k Keeper) GetAccountAsBytes(address string) ([]byte, error) {
	return k.GetAddressCodec().StringToBytes(address)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetState returns the current state of the bond module.
func (k Keeper) GetState(ctx context.Context) types.State {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.StateKey)
	if err != nil || bz == nil {
		return types.DefaultState()
	}
	var state types.State
	k.cdc.MustUnmarshal(bz, &state)
	return state
}

// SetState sets the state of the bond module.
func (k Keeper) SetState(ctx context.Context, state types.State) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&state)
	return store.Set(types.StateKey, bz)
}

// GetYieldMintHeight returns the height at which the yield was minted.
func (k Keeper) GetYieldMintHeight(ctx context.Context) int64 {
	return k.GetState(ctx).YieldMintHeight
}

// SetYieldMintHeight sets the height at which the yield was minted.
func (k Keeper) SetYieldMintHeight(ctx context.Context, height int64) error {
	state := k.GetState(ctx)
	state.YieldMintHeight = height
	return k.SetState(ctx, state)
}
