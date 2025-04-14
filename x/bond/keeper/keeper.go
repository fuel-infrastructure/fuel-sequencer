package keeper

import (
	"context"
	"fmt"

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
		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	bankKeeper types.BankKeeper,
) Keeper {
	if authority != "" {
		if _, err := sdk.AccAddressFromBech32(authority); err != nil {
			panic(err)
		}
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
		bankKeeper:   bankKeeper,
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

// BurnCoins burns coins from the sender's account
func (k Keeper) BurnCoins(ctx context.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	// Validate coins
	if !coins.IsValid() {
		return fmt.Errorf("invalid coins: %s", coins)
	}

	if coins.IsZero() {
		return fmt.Errorf("coins cannot be zero")
	}

	// Send coins from account to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to send coins to module: %w", err)
	}

	// Burn the coins
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to burn coins: %w", err)
	}

	return nil
}

// GetBankKeeper returns the bank keeper
func (k Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

// GetCodec returns the codec
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// GetStoreService returns the store service
func (k Keeper) GetStoreService() store.KVStoreService {
	return k.storeService
}
