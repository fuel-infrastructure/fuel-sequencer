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
		height        int64
		delAddr       sdk.AccAddress
		valAddr       sdk.ValAddress
		slashAmount   sdkmath.Int
		expSlashEntry types.SlashEntry
		expErrMsg     string
		expPanic      bool
	}{
		// There is no need to test all cases here because InsertSlashEntry has its own unit tests

		{
			name:        "slash entry set correctly if no errors from InsertSlashEntry - no existent slash entry",
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
			name:        "panics if InsertSlashEntry errors - context height is zero",
			height:      0,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg: "AfterUnbondingDelegationSlashed: could not insert slash entry: slashing height must be" +
				" positive, received: 0",
			expPanic: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Update context's height as required by test
			ctxWithHeight := s.Ctx().WithBlockHeight(tc.height)

			// Execute hook
			if tc.expPanic {

				// Expect a panic if required by test
				var panicMsg string
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicMsg = r.(error).Error()
						}
					}()

					err := s.App.ReportsKeeper.Hooks().AfterUnbondingDelegationSlashed(
						ctxWithHeight, tc.valAddr, tc.delAddr, tc.slashAmount,
					)
					s.Require().Fail(fmt.Sprintf("Expected panic but got error: %v", err))
				}()
				s.Require().Contains(panicMsg, tc.expErrMsg)

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

func (s *KeeperTestSuite) TestAfterRedelegationSlashed() {
	height := int64(10)
	delegatorAccAddress := testtypes.TestSeqAddr1
	validatorAddress, err := sdk.ValAddressFromBech32(testtypes.TestValAddr1Str)
	s.Require().NoError(err)

	testCases := []struct {
		name          string
		height        int64
		delAddr       sdk.AccAddress
		valAddr       sdk.ValAddress
		slashAmount   sdkmath.Int
		expSlashEntry types.SlashEntry
		expErrMsg     string
		expPanic      bool
	}{
		// There is no need to test all cases here because InsertSlashEntry has its own unit tests

		{
			name:        "slash entry set correctly if no errors from InsertSlashEntry - no existent slash entry",
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
			name:        "panics if InsertSlashEntry errors - context height is zero",
			height:      0,
			delAddr:     delegatorAccAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg: "AfterRedelegationSlashed: could not insert slash entry: slashing height must be positive, " +
				"received: 0",
			expPanic: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Update context's height as required by test
			ctxWithHeight := s.Ctx().WithBlockHeight(tc.height)

			// Execute hook
			if tc.expPanic {

				// Expect a panic if required by test
				var panicMsg string
				func() {
					defer func() {
						if r := recover(); r != nil {
							panicMsg = r.(error).Error()
						}
					}()

					err := s.App.ReportsKeeper.Hooks().AfterRedelegationSlashed(
						ctxWithHeight, tc.valAddr, tc.delAddr, tc.slashAmount,
					)
					s.Require().Fail(fmt.Sprintf("Expected panic but got error: %v", err))
				}()
				s.Require().Contains(panicMsg, tc.expErrMsg)

			} else {

				// If no panic is expected, run the hook without wrapping a deferred function and check for errors or
				// the execution results, as required by the test case.
				err := s.App.ReportsKeeper.Hooks().AfterRedelegationSlashed(
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
