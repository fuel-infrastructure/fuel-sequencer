package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func (s *KeeperTestSuite) TestAfterUnbondingDelegationSlashed() {
	height := int64(10)
	delegatorAccAddress := testtypes.TestSeqAddr1
	validatorAddress, err := sdk.ValAddressFromBech32(testtypes.TestValAddr1Str)
	s.Require().NoError(err)

	testCases := []struct {
		name          string
		setSlashEntry *types.SlashEntry
		height        int64
		delAddr       sdk.AccAddress
		valAddr       sdk.ValAddress
		slashAmount   sdkmath.Int
		expSlashEntry types.SlashEntry
		expErrMsg     string
		expPanic      bool
	}{
		{
			name:        "slash entry set correctly first time",
			height:      height,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expSlashEntry: types.SlashEntry{
				ValidatorAddress:          testtypes.TestValAddr1Str,
				DelegatorAddress:          testtypes.TestSeqAddr1Str,
				DelegatorSlashAmount:      sdkmath.OneInt(),
				DelegatorBondedBalance:    sdkmath.ZeroInt(),
				DelegatorUnbondingBalance: sdkmath.ZeroInt(),
			},
		},
		{
			name: "slash entry updated correctly if already exists",
			setSlashEntry: &types.SlashEntry{
				ValidatorAddress:          testtypes.TestValAddr1Str,
				DelegatorAddress:          testtypes.TestSeqAddr1Str,
				DelegatorSlashAmount:      sdkmath.OneInt(),
				DelegatorBondedBalance:    sdkmath.OneInt(), // Set to one to make sure that the hook updates value to 0
				DelegatorUnbondingBalance: sdkmath.OneInt(), // Set to one to make sure that the hook updates value to 0
			},
			height:      height,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expSlashEntry: types.SlashEntry{
				ValidatorAddress:          testtypes.TestValAddr1Str,
				DelegatorAddress:          testtypes.TestSeqAddr1Str,
				DelegatorSlashAmount:      sdkmath.NewInt(2), // 1 (existing entry) + 1 (new slash amount)
				DelegatorBondedBalance:    sdkmath.ZeroInt(),
				DelegatorUnbondingBalance: sdkmath.ZeroInt(),
			},
		},
		{
			name:        "panics if context height is zero",
			height:      0,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg:   "slashing height must be positive, received: 0",
			expPanic:    true,
		},
		{
			name:        "panics if context height is negative",
			height:      -1,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg:   "slashing height must be positive, received: -1",
			expPanic:    true,
		},
		{
			name:        "panics if slashed amount is zero",
			height:      height,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.ZeroInt(),
			expErrMsg:   "slash amount must be positive, received 0",
			expPanic:    true,
		},
		{
			name:        "panics if slashed amount is negative",
			height:      height,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.NewInt(-1),
			expErrMsg:   "slash amount must be positive, received -1",
			expPanic:    true,
		},
		// The test case below was only written for the sake of completion. We should never be in a situation where a
		// slash entry is constructed in an invalid manner. In fact to trigger this case we will be storing a negative
		// slash amount in state, which should never occur if we call validate basic before storing.
		{
			name: "panics if constructed invalid slash entry",
			setSlashEntry: &types.SlashEntry{
				ValidatorAddress: testtypes.TestValAddr1Str,
				DelegatorAddress: testtypes.TestSeqAddr1Str,

				// Set to a value > slash amount so resultant slash entry has negative slash amount
				DelegatorSlashAmount: sdkmath.NewInt(-10),

				DelegatorBondedBalance:    sdkmath.ZeroInt(),
				DelegatorUnbondingBalance: sdkmath.ZeroInt(),
			},
			height:      height,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.NewInt(-1),
			expErrMsg:   "AfterUnbondingDelegationSlashed: constructed invalid slash entry",
			expPanic:    true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Update context's height as required by test
			ctxWithHeight := s.Ctx().WithBlockHeight(tc.height)

			// Set slash entry if required by test
			if tc.setSlashEntry != nil {
				s.App.ReportsKeeper.SetSlashEntry(ctxWithHeight, uint64(tc.height), *tc.setSlashEntry)
			}

			// Execute hook
			if tc.expPanic {

				// Expect a panic if required by test
				s.Require().Panics(func() {
					defer func() {
						if r := recover(); r != nil {
							s.Require().Contains(r.(string), tc.expErrMsg)
						}
					}()

					err := s.App.ReportsKeeper.Hooks().AfterUnbondingDelegationSlashed(
						ctxWithHeight, tc.valAddr, tc.delAddr, tc.slashAmount,
					)
					s.Require().Fail(fmt.Sprintf("Expected panic but got error: %v", err))
				})
			} else {

				// If no panic is expected, run the hook without wrapping a deferred function and check for errors or
				// the execution results, as required by the test case.
				err := s.App.ReportsKeeper.Hooks().AfterUnbondingDelegationSlashed(
					ctxWithHeight, tc.valAddr, tc.delAddr, tc.slashAmount,
				)
				if len(tc.expErrMsg) > 0 {
					s.Require().Error(err)
					s.Require().ErrorContains(err, tc.expErrMsg)
					return
				}
				s.Require().NoError(err)

				actualSlashEntry, found := s.App.ReportsKeeper.GetSlashEntry(
					ctxWithHeight, uint64(tc.height), tc.delAddr.String(), tc.valAddr.String(),
				)
				s.Require().True(found)
				s.Require().Equal(tc.expSlashEntry, actualSlashEntry)
			}
		})
	}
}
