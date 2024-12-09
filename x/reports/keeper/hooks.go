package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Make sure that the height is positive
	// Given that slashing is not expected at heights less or equal to zero we will panic if this occurs.
	if sdkCtx.BlockHeight() <= 0 {
		panic(sdkerrors.ErrInvalidHeight.Wrapf("slashing height must be positive, received: %v", sdkCtx.BlockHeight()))
	}

	// Compute height, validator address and delegator address.
	// Note: uint64 is bigger than int64, therefore, the height can be converted safely to uint64.
	height := uint64(sdkCtx.BlockHeight())
	validatorAddress := valAddr.String()
	delegatorAddress := delAddr.String()

	// Make sure that slashed amount is positive. We are panicking here because this is highly unexpected, as the forked
	// Cosmos SDK triggers the AfterUnbondingDelegationSlashed hook only if the burned amount is positive.
	if !slashAmount.IsPositive() {
		panic(fmt.Errorf("slash amount must be positive, received %s", slashAmount.String()))
	}

	// Check if a slash entry already exists in state for the current height.
	slashEntry, found := h.k.GetSlashEntry(ctx, height, delegatorAddress, validatorAddress)
	if found {
		// If a slash entry already exists, increment the slashed amount as it must be that the delegator has already
		// been slashed at this height. Set the unbonding and bonded balances to zero as that should be computed by the
		// reports module BeginBlocker.
		slashEntry.DelegatorSlashAmount = slashEntry.DelegatorSlashAmount.Add(slashAmount)
		slashEntry.DelegatorBondedBalance = sdkmath.ZeroInt()
		slashEntry.DelegatorUnbondingBalance = sdkmath.ZeroInt()
	} else {
		// Otherwise create a new slash entry. For the same reason as the found=true case, delegator bonded and
		// unbonding balances should be set to zero.
		slashEntry = types.SlashEntry{
			ValidatorAddress:          validatorAddress,
			DelegatorAddress:          delegatorAddress,
			DelegatorSlashAmount:      slashAmount,
			DelegatorBondedBalance:    sdkmath.ZeroInt(),
			DelegatorUnbondingBalance: sdkmath.ZeroInt(),
		}
	}

	// Make sure that the slash entry satisfies the basic validation checks and panic if not as this would mean that
	// there is a bug in the code. We need to panic because the staking module delegates panicking responsibilities to
	// the individual modules.
	if err := slashEntry.ValidateBasic(); err != nil {
		panic(errors.Wrap(err, "AfterUnbondingDelegationSlashed: constructed invalid slash entry"))
	}

	h.k.SetSlashEntry(ctx, height, slashEntry)

	return nil
}
