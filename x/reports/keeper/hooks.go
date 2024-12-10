package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

var _ types.StakingHooks = Hooks{}

// Hooks wrapper struct for reports keeper
type Hooks struct {
	k Keeper
}

// Hooks returns the reports module hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorSlashed(_ context.Context, _ sdk.ValAddress, _ sdkmath.LegacyDec) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

func (h Hooks) AfterUnbondingDelegationSlashed(
	ctx context.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, slashAmount sdkmath.Int,
) error {
	// Create slash entry for the unbonding delegation
	if err := h.k.InsertSlashEntry(ctx, valAddr, delAddr, slashAmount); err != nil {

		// The function will panic if an error is returned, as any error from InsertSlashEntry is highly unlikely
		// and would indicate a logical error in the code.
		// Note: The staking module does not panic when calling AfterUnbondingDelegationSlashed. Therefore, we must
		// panic here to ensure the chain is halted.
		panic(errors.Wrap(err, "AfterUnbondingDelegationSlashed: could not insert slash entry"))
	}

	return nil
}

func (h Hooks) AfterRedelegationSlashed(
	ctx context.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, slashAmount sdkmath.Int,
) error {
	// Create slash entry for the redelegation
	if err := h.k.InsertSlashEntry(ctx, valAddr, delAddr, slashAmount); err != nil {

		// The function will panic if an error is returned, as any error from InsertSlashEntry is highly unlikely
		// and would indicate a logical error in the code.
		// Note: The staking module does not panic when calling AfterRedelegationSlashed. Therefore, we must panic here
		// to ensure the chain is halted.
		panic(errors.Wrap(err, "AfterRedelegationSlashed: could not insert slash entry"))
	}

	return nil
}
