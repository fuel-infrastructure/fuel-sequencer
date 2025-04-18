package token_updates

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

const UpgradeName = "token-updates"

const vestingPeriodTrim = int64(15768000) // 6 months

var accForUpgrade = []string{}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

		sdkCtx := types.UnwrapSDKContext(ctx)
		now := sdkCtx.BlockTime().Unix()

		for _, accStr := range accForUpgrade {

			sdkAcc, err := accountKeeper.AddressCodec().StringToBytes(accStr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert account %s to bytes: %w", accStr, err)
			}

			acc := accountKeeper.GetAccount(ctx, sdkAcc)
			if acc == nil {
				return nil, fmt.Errorf("account %s not found", accStr)
			}

			vestingAcc, ok := acc.(*bridgetypes.EthOwnedContinuousVestingAccount)
			if !ok {
				return nil, fmt.Errorf("account %s is not a continuous vesting account", accStr)
			}

			// TODO: note that tokens that haven't been migrated yet won't benefit from this trim

			newVestingEndTime := vestingAcc.EndTime - vestingPeriodTrim
			if newVestingEndTime <= now {
				// Trim is greater than the remaining vesting duration, so all tokens get unlocked.
				vestingAcc.EndTime = now
			} else {
				// Calculate percentage of vesting time being trimmed.
				totalVestingTime := vestingAcc.EndTime - vestingAcc.StartTime
				trimPercentage := math.LegacyNewDec(vestingPeriodTrim).QuoInt64(totalVestingTime)
				if trimPercentage.GT(math.LegacyOneDec()) {
					return nil, fmt.Errorf("trim is greater than total vesting time for account %s", accStr)
				}

				// Calculate tokens getting unlocked early.
				earlyUnlockDec := types.NewDecCoinsFromCoins(vestingAcc.OriginalVesting...).MulDec(trimPercentage)
				earlyUnlock, _ := earlyUnlockDec.TruncateDecimal()

				// Unlock tokens.
				vestingAcc.OriginalVesting = vestingAcc.OriginalVesting.Sub(earlyUnlock...)
				vestingAcc.EndTime = newVestingEndTime
			}

			if err := vestingAcc.Validate(); err != nil {
				return nil, err
			}
		}

		// returns a VersionMap with the updated module ConsensusVersions
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
