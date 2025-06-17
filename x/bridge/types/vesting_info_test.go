package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestVestingInfo_GetVestedCoins(t *testing.T) {
	now := time.Now()
	endTime := now.Add(24 * time.Hour)

	coins := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	halfCoins := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))

	info := types.NewVestingInfo(coins, now.Unix(), endTime.Unix())

	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := info.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = info.GetVestedCoins(endTime)
	require.Equal(t, coins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = info.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, halfCoins, vestedCoins)

	// require 100% of coins vested
	vestedCoins = info.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, coins, vestedCoins)
}

func TestVestingInfo_Validate(t *testing.T) {

	coins := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "_invalid_denom_", Amount: sdkmath.NewInt(100)}}
	negativeCoins := sdk.Coins{sdk.Coin{Denom: testutiltypes.TestToken, Amount: sdkmath.NewInt(-1)}}
	emptyCoins := sdk.NewCoins()
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")

	testCases := []struct {
		name      string
		coins     sdk.Coins
		startTime int64
		endTime   int64
		expErr    bool
	}{
		{"valid", coins, t0.Unix(), t1.Unix(), false},
		{"end time right after start time", coins, t0.Unix(), t0.Unix() + 1, false},
		{"start time == end time", coins, t0.Unix(), t0.Unix(), true},
		{"negative start time", coins, -1, t1.Unix(), true},
		{"negative end time", coins, t0.Unix(), -1, true},
		{"invalid denom coins", invalidCoins, t0.Unix(), t1.Unix(), true},
		{"invalid negative coins", negativeCoins, t0.Unix(), t1.Unix(), true},
		{"invalid empty coins", emptyCoins, t0.Unix(), t1.Unix(), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expErr {
				require.Error(t, types.NewVestingInfo(tc.coins, tc.startTime, tc.endTime).Validate())
			} else {
				require.NoError(t, types.NewVestingInfo(tc.coins, tc.startTime, tc.endTime).Validate())
			}
		})
	}
}
