package types

import (
	"context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	AddressCodec() address.Codec

	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)

	GetModulePermissions() map[string]authtypes.PermissionsForAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	GetSupply(ctx context.Context, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(
		ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(
		ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// StakingKeeper defines the expected interface for the Staking module.
type StakingKeeper interface {
	// GetLastValidators gets the bonded validators
	GetLastValidators(ctx context.Context) ([]stakingtypes.Validator, error)
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
