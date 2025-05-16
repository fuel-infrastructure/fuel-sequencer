package types

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// AccountKeeper defines the expected interface for the Account module.
// Subset of authkeeper.AccountKeeperI
type AccountKeeper interface {
	// AddressCodec returns the account address codec.
	AddressCodec() address.Codec

	// GetModuleAccount returns the module account for the given module name.
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// BridgeKeeper defines the contract needed to be fulfilled for bridge module dependencies.
type BridgeKeeper interface {
	GetParams(ctx context.Context) bridgetypes.Params
}
