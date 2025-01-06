package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}

type StakingHooks interface {
	AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error
	BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error
	AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error

	AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error
	AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error

	BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error

	// These are custom hooks
	AfterUnbondingDelegationSlashed(
		ctx context.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, slashedAmount math.Int,
	) error
	AfterRedelegationSlashed(
		ctx context.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, slashedAmount math.Int,
	) error
	CustomBeforeValidatorSlashed(
		ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec, totalSlashedAmt math.Int,
	) error
}

// StakingKeeper defines the expected interface for the Staking module.
type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)

	IterateValidatorDelegations(
		ctx context.Context, valAddr sdk.ValAddress, cb func(delegation stakingtypes.Delegation) (stop bool),
	) error

	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error)

	GetUnbondingDelegation(
		ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	) (ubd stakingtypes.UnbondingDelegation, err error)

	// Methods imported from x/staking should be defined here
}
