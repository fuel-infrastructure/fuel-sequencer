package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
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

func (s *KeeperTestSuite) TestCustomBeforeValidatorSlashed() {
	height := int64(10)

	validatorAddress, err := sdk.ValAddressFromBech32(testtypes.TestValAddr1Str)
	s.Require().NoError(err)

	testDelegators := sims.CreateRandomAccounts(3)

	testCases := []struct {
		name            string
		height          int64
		setValidator    *stakingtypes.Validator
		setDelegations  *[]stakingtypes.Delegation
		valAddr         sdk.ValAddress
		fraction        sdkmath.LegacyDec
		totalSlashedAmt sdkmath.Int
		expSlashEntries []types.SlashEntry
		expErrMsg       string
		expPanic        bool
	}{
		// There is no need to test all cases related to InsertSlashEntry here because InsertSlashEntry has its own unit
		// tests

		{
			name:   "slash entries set correctly if everything valid - shares equal to number of tokens",
			height: height,

			// This is the validator object before slash
			setValidator: &stakingtypes.Validator{
				OperatorAddress:         validatorAddress.String(),
				ConsensusPubkey:         nil,
				Jailed:                  false,
				Status:                  0,
				Tokens:                  sdkmath.NewInt(30),                    // 10 tokens per delegator
				DelegatorShares:         sdkmath.LegacyMustNewDecFromStr("30"), // 1 token per share
				Description:             stakingtypes.Description{},
				UnbondingHeight:         0,
				UnbondingTime:           time.Time{},
				Commission:              stakingtypes.Commission{},
				MinSelfDelegation:       sdkmath.Int{},
				UnbondingOnHoldRefCount: 0,
				UnbondingIds:            nil,
			},

			// These are the delegations before slash
			setDelegations: &[]stakingtypes.Delegation{
				{
					DelegatorAddress: testDelegators[0].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("10"),
				},
				{
					DelegatorAddress: testDelegators[1].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("10"),
				},
				{
					DelegatorAddress: testDelegators[2].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("10"),
				},
			},
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.1"), // 10% effective slash per delegator
			totalSlashedAmt: sdkmath.OneInt(),                       // value irrelevant as long as it is positive
			expSlashEntries: []types.SlashEntry{
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[0].String(),

					// slashAmount = 10 * 0.1 = 1 token
					DelegatorSlashAmount:      sdkmath.OneInt(),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[1].String(),

					// slashAmount = 10 * 0.1 = 1 token
					DelegatorSlashAmount:      sdkmath.OneInt(),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[2].String(),

					// slashAmount = 10 * 0.1 = 1 token
					DelegatorSlashAmount:      sdkmath.OneInt(),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
			},
		},
		{
			name:   "slash entries set correctly if everything valid - shares not equal to number of tokens",
			height: height,

			// This is the validator object before slash
			setValidator: &stakingtypes.Validator{
				OperatorAddress:         validatorAddress.String(),
				ConsensusPubkey:         nil,
				Jailed:                  false,
				Status:                  0,
				Tokens:                  sdkmath.NewInt(1000),                  // shares not 1:1 with number of tokens
				DelegatorShares:         sdkmath.LegacyMustNewDecFromStr("33"), // total shares across 3 delegators
				Description:             stakingtypes.Description{},
				UnbondingHeight:         0,
				UnbondingTime:           time.Time{},
				Commission:              stakingtypes.Commission{},
				MinSelfDelegation:       sdkmath.Int{},
				UnbondingOnHoldRefCount: 0,
				UnbondingIds:            nil,
			},

			// These are the delegations before slash
			setDelegations: &[]stakingtypes.Delegation{
				{
					DelegatorAddress: testDelegators[0].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("5"),
				},
				{
					DelegatorAddress: testDelegators[1].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("20"),
				},
				{
					DelegatorAddress: testDelegators[2].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("8"),
				},
			},
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.15"), // 15% effective slash per delegator
			totalSlashedAmt: sdkmath.OneInt(),                        // value irrelevant as long as it is positive
			expSlashEntries: []types.SlashEntry{
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[0].String(),

					// currentTokens = (5 * 1000) / 33 = 151.515151515151515151
					// slashAmount = 151.515151515151515151 * 0.15 = 22.727272727272727272 = 22 tokens
					DelegatorSlashAmount:      sdkmath.NewInt(22),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[1].String(),

					// currentTokens = (20 * 1000) / 33 = 606.060606060606060606
					// slashAmount = 606.060606060606060606 * 0.15 = 90.909090909090909090 = 90 tokens
					DelegatorSlashAmount:      sdkmath.NewInt(90),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
				{
					ValidatorAddress: validatorAddress.String(),
					DelegatorAddress: testDelegators[2].String(),

					// currentTokens = (8 * 1000) / 33 = 242.424242424242424242
					// slashAmount = 242.424242424242424242 * 0.15 = 36.363636363636363636 = 36 tokens
					DelegatorSlashAmount:      sdkmath.NewInt(36),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
			},
		},
		{
			name:            "panics if total slashed amount is zero",
			height:          height,
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.1"),
			totalSlashedAmt: sdkmath.ZeroInt(),
			expErrMsg: fmt.Sprintf(
				"CustomBeforeValidatorSlashed: total slashed amount must be positive, received: %v",
				sdkmath.ZeroInt(),
			),
			expPanic: true,
		},
		{
			name:            "panics if total slashed amount is negative",
			height:          height,
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.1"),
			totalSlashedAmt: sdkmath.NewInt(-1),
			expErrMsg: fmt.Sprintf(
				"CustomBeforeValidatorSlashed: total slashed amount must be positive, received: %v",
				sdkmath.NewInt(-1),
			),
			expPanic: true,
		},
		{
			name:            "panics if fraction is greater than one",
			height:          height,
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("2"),
			totalSlashedAmt: sdkmath.OneInt(),
			expErrMsg: fmt.Sprintf(
				"CustomBeforeValidatorSlashed: fraction must be >0 and <=1, current fraction: %v",
				sdkmath.LegacyMustNewDecFromStr("2"),
			),
			expPanic: true,
		},
		{
			name:            "panics if fraction is zero",
			height:          height,
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyZeroDec(),
			totalSlashedAmt: sdkmath.OneInt(),
			expErrMsg: fmt.Sprintf(
				"CustomBeforeValidatorSlashed: fraction must be >0 and <=1, current fraction: %v",
				sdkmath.LegacyZeroDec(),
			),
			expPanic: true,
		},
		{
			name:            "panics if fraction is negative",
			height:          height,
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("-1"),
			totalSlashedAmt: sdkmath.OneInt(),
			expErrMsg: fmt.Sprintf(
				"CustomBeforeValidatorSlashed: fraction must be >0 and <=1, current fraction: %v",
				sdkmath.LegacyMustNewDecFromStr("-1"),
			),
			expPanic: true,
		},
		{
			name:   "panics if could not find validator object",
			height: height,
			setDelegations: &[]stakingtypes.Delegation{
				{
					DelegatorAddress: testDelegators[1].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("10"),
				},
			},
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.1"),
			totalSlashedAmt: sdkmath.OneInt(),
			expErrMsg:       "CustomBeforeValidatorSlashed: could not get validator",
			expPanic:        true,
		},
		{
			name:   "panics if InsertSlashEntry errors - context height is zero",
			height: 0,
			setValidator: &stakingtypes.Validator{
				OperatorAddress:         validatorAddress.String(),
				ConsensusPubkey:         nil,
				Jailed:                  false,
				Status:                  0,
				Tokens:                  sdkmath.NewInt(30),
				DelegatorShares:         sdkmath.LegacyMustNewDecFromStr("30"),
				Description:             stakingtypes.Description{},
				UnbondingHeight:         0,
				UnbondingTime:           time.Time{},
				Commission:              stakingtypes.Commission{},
				MinSelfDelegation:       sdkmath.Int{},
				UnbondingOnHoldRefCount: 0,
				UnbondingIds:            nil,
			},
			setDelegations: &[]stakingtypes.Delegation{
				// One delegator is enough to trigger the error
				{
					DelegatorAddress: testDelegators[1].String(),
					ValidatorAddress: validatorAddress.String(),
					Shares:           sdkmath.LegacyMustNewDecFromStr("10"),
				},
			},
			valAddr:         validatorAddress,
			fraction:        sdkmath.LegacyMustNewDecFromStr("0.1"),
			totalSlashedAmt: sdkmath.OneInt(),
			expErrMsg: "CustomBeforeValidatorSlashed: could not insert slash entry: slashing height must be " +
				"positive, received: 0",
			expPanic: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Update context's height as required by test
			ctxWithHeight := s.Ctx().WithBlockHeight(tc.height)

			// Set validator in state if required by test
			if tc.setValidator != nil {
				err := s.App.StakingKeeper.SetValidator(ctxWithHeight, *tc.setValidator)
				s.Require().NoError(err)
			}

			// Set delegations in state if required by test
			if tc.setDelegations != nil {
				for _, delegation := range *tc.setDelegations {
					err := s.App.StakingKeeper.SetDelegation(ctxWithHeight, delegation)
					s.Require().NoError(err)
				}
			}

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

					err := s.App.ReportsKeeper.Hooks().CustomBeforeValidatorSlashed(
						ctxWithHeight, tc.valAddr, tc.fraction, tc.totalSlashedAmt,
					)
					s.Require().Fail(fmt.Sprintf("Expected panic but got error: %v", err))
				}()
				s.Require().Contains(panicMsg, tc.expErrMsg)

			} else {

				// If no panic is expected, run the hook without wrapping a deferred function and check for errors or
				// the execution results, as required by the test case.
				err := s.App.ReportsKeeper.Hooks().CustomBeforeValidatorSlashed(
					ctxWithHeight, tc.valAddr, tc.fraction, tc.totalSlashedAmt,
				)
				if len(tc.expErrMsg) > 0 {
					s.Require().Error(err)
					s.Require().ErrorContains(err, tc.expErrMsg)
					return
				}
				s.Require().NoError(err)

				slashReport, found := s.App.ReportsKeeper.GetSlashReport(ctxWithHeight, uint64(tc.height))
				s.Require().True(found)
				s.Require().Equal(
					keepertest.OrderSlashEntriesLexicographically(tc.expSlashEntries),
					keepertest.OrderSlashEntriesLexicographically(slashReport.Entries),
				)
			}
		})
	}
}
