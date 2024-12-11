package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (h Hooks) AfterValidatorSlashed(
	ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec, valSlashedAmt sdkmath.Int,
) error {

	// If the validator slash amount is not positive, there is likely an issue with the staking module's Slash function,
	// as this hook should only be called when valSlashedAmt is positive. We opt for a panic, prioritizing safety over
	// liveness.
	if !valSlashedAmt.IsPositive() {
		panic(fmt.Errorf(
			"AfterValidatorSlashed: validator slashed amount must be positive, received: %s", valSlashedAmt.String(),
		))
	}

	// If the fraction is less than or equal to 0, or greater than 1, there is likely an issue with the staking module's
	// Slash function. In this case, we must panic because tokens have been burned, but we are unable to compute the
	// portion that was slashed for the delegators.
	if !fraction.IsPositive() || fraction.GT(sdkmath.LegacyOneDec()) {
		panic(fmt.Errorf("AfterValidatorSlashed: fraction must be >0 and <=1, current fraction: %v", fraction))
	}

	// At this stage, we are certain that redelegations and unbonding delegations that were active at the time of the
	// infraction have already been slashed. Therefore, we can iterate over active delegations and calculate the slashed
	// amount.
	// Note: This is only possible because the fraction represents the percentage of tokens slashed, excluding
	// redelegations and unbonding delegations that were active at the time of the infraction.
	err := h.k.stakingKeeper.IterateValidatorDelegations(
		ctx, valAddr, func(delegation stakingtypes.Delegation) (stop bool) {
			delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

			// For every delegation, get the validator object from state. There must be something really wrong if the
			// validator object could not be obtained, therefore, in that case panic.
			validator, err := h.k.stakingKeeper.GetValidator(ctx, valAddr)
			if err != nil {
				panic(errors.Wrap(err, "AfterValidatorSlashed: could not get validator"))
			}

			// tokensBeforeSlash = (currentTokens) / (1-effectiveFraction)
			// We perform the calculations on Dec to be as precise as possible when reversing the slash.
			delegatorCurrentTokensDec := validator.TokensFromShares(delegation.GetShares())
			divisor := sdkmath.LegacyOneDec().Sub(fraction)
			delegatorTokensBeforeDec := delegatorCurrentTokensDec.Quo(divisor)

			// We truncate the decimals to mirror the staking module.
			delegatorCurrentTokens := delegatorCurrentTokensDec.TruncateInt()
			delegatorTokensBefore := delegatorTokensBeforeDec.TruncateInt()
			delSlashAmt := delegatorTokensBefore.Sub(delegatorCurrentTokens)

			// Skip if slashed amount is zero as this means that the effective fraction is so low that the slash is
			// negligible. In other words this means that the delegator was not slashed.
			// Notes:
			// 1. This was done for the sake of completion and is unexpected in 99.99% of the cases.
			// 2. If slashed amount is negative we want to panic as this means that something went wrong in the
			// calculation. For this case we are relying on the fact that InsertSlashEntry errors if the slash amount is
			// not positive.
			if delSlashAmt.IsZero() {
				return false
			}

			// Create slash entry for the delegator
			if err := h.k.InsertSlashEntry(ctx, valAddr, delAddr, delSlashAmt); err != nil {

				// The function will panic if an error is returned, as any error from InsertSlashEntry is highly
				// unlikely and would indicate a logical error in the code.
				// Note: The staking module does not panic when calling AfterValidatorSlashed. Therefore, we must panic
				// here to ensure the chain is halted.
				panic(errors.Wrap(err, "AfterValidatorSlashed: could not insert slash entry"))
			}

			return false
		},
	)
	if err != nil {

		// If we error while iterating over delegations there is something wrong in the store. Therefore, panic since we
		// could not store the slash entries.
		panic(errors.Wrapf(err, "AfterValidatorSlashed: could not iterate validator %s delegations", valAddr.String()))
	}

	return nil
}
