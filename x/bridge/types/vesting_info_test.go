package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestVestingInfo_GetVestedCoins(t *testing.T) {
	now := time.Now()
	endTime := now.Add(24 * time.Hour)

	coins := sdk.NewCoins(sdk.NewInt64Coin(types.DefaultBridgeDenom, 100))
	halfCoins := sdk.NewCoins(sdk.NewInt64Coin(types.DefaultBridgeDenom, 50))

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
