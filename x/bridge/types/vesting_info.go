package types

import (
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVestingInfo(originalVesting sdk.Coins, startTime, endTime int64) *VestingInfo {
	return &VestingInfo{
		OriginalVesting: originalVesting,
		StartTime:       startTime,
		EndTime:         endTime,
	}
}

// GetVestedCoins replicates the ContinuousVestingAccount's GetVestedCoins function since this is the behaviour we need.
// https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L198
func (vi *VestingInfo) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= vi.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= vi.EndTime {
		return vi.OriginalVesting
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - vi.StartTime
	y := vi.EndTime - vi.StartTime
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	for _, ovc := range vi.OriginalVesting {
		vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}

// Validate somewhat replicates the ContinuousVestingAccount's Validate function.
// https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L249
func (vi *VestingInfo) Validate() error {
	if vi.StartTime < 0 {
		return errors.New("start time cannot be negative")
	}

	if vi.GetStartTime() >= vi.GetEndTime() {
		return errors.New("vesting start-time cannot be before end-time")
	}

	if !vi.OriginalVesting.IsValid() || !vi.OriginalVesting.IsAllPositive() {
		return fmt.Errorf("invalid coins: %s", vi.OriginalVesting.String())
	}

	return nil
}
