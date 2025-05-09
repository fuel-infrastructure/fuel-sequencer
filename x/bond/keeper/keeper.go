package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

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
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
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

// AddCollectedBondAllocation transfers bond allocation coins from the mint module to the bond authority.
// This function assumes that the mint module has already minted the specified amount of tokens
// specifically for this bond allocation.
// AddCollectedBondAllocation to be used in custom x/mint module's BeginBlocker.
func (k Keeper) AddCollectedBondAllocation(ctx context.Context, allocation sdk.Coins) error {
	if allocation.IsZero() {
		return nil
	}

	// Get the bond authority address
	bondAuthority, err := k.GetAccountAsBytes(k.authority)
	if err != nil {
		return err
	}
	// Send coins from mint module to bond authority
	return k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, minttypes.ModuleName, bondAuthority, allocation,
	)
}
