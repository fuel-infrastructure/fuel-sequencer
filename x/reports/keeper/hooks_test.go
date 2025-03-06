package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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

func (s *KeeperTestSuite) TestAfterUnbondingDelegationSlashed_Integration() {

	currHeight := int64(100)
	infractionHeight := currHeight
	currTime := time.Now()
	hctx := s.Ctx().WithBlockHeight(currHeight).WithBlockTime(currTime)
	slashFraction := sdkmath.LegacyOneDec()
	undelegationBalance := sdkmath.NewIntWithDecimal(1, 9)

	// Sanity check Governance account balance pre-slash
	govAccount := s.App.GovKeeper.GetGovernanceAccount(hctx).GetAddress()
	govAccountBalance := s.App.BankKeeper.GetBalance(hctx, govAccount, sdk.DefaultBondDenom)
	s.Require().True(govAccountBalance.IsZero())

	// Sanity check the NotBondedPoolName balance
	acc := s.App.AccountKeeper.GetModuleAccount(hctx, stakingtypes.NotBondedPoolName)
	s.Require().Empty(s.App.BankKeeper.GetAllBalances(hctx, acc.GetAddress()))

	// Mint the amount to be burned to the NotBondedPoolName, to avoid errors.
	expectedBurn := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, undelegationBalance))
	err := s.App.BankKeeper.MintCoins(hctx, minttypes.ModuleName, expectedBurn)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(
		hctx, minttypes.ModuleName, stakingtypes.NotBondedPoolName, expectedBurn,
	)
	s.Require().NoError(err)
	s.Require().Equal(expectedBurn, s.App.BankKeeper.GetAllBalances(hctx, acc.GetAddress()))

	// Unbonding delegation
	undelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: testtypes.TestSeqAddr1Str, // arbitrary
		ValidatorAddress: testtypes.TestValAddr1Str, // arbitrary
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CreationHeight:          currHeight + 1,                   // after current height
				CompletionTime:          currTime.Add(time.Hour),          // after current time
				InitialBalance:          sdkmath.NewIntWithDecimal(1, 20), // Should be >= Balance
				Balance:                 undelegationBalance,              // Amount that will get slashed
				UnbondingId:             0,                                // n/a
				UnbondingOnHoldRefCount: 0,                                // n/a
			},
		},
	}

	_, err = s.App.StakingKeeper.SlashUnbondingDelegation(hctx, undelegation, infractionHeight, slashFraction)
	s.Require().NoError(err)

	slashReport, found := s.App.ReportsKeeper.GetSlashReport(hctx, uint64(currHeight))
	s.Require().True(found)

	expectedTotalSlash := undelegationBalance // all tokens slashed
	expectedSlashReport := types.SlashReport{
		Height: uint64(currHeight),
		Entries: []types.SlashEntry{
			{
				ValidatorAddress:          testtypes.TestValAddr1Str,
				DelegatorAddress:          testtypes.TestSeqAddr1Str,
				DelegatorSlashAmount:      expectedTotalSlash,
				DelegatorBondedBalance:    sdkmath.ZeroInt(), // n/a
				DelegatorUnbondingBalance: sdkmath.ZeroInt(), // n/a
			},
		},
	}
	s.Require().Equal(expectedSlashReport, slashReport)

	// Sanity check that tokens don't actually get burned
	govAccountBalance = s.App.BankKeeper.GetBalance(hctx, govAccount, sdk.DefaultBondDenom)
	s.Require().True(govAccountBalance.Amount.Equal(expectedTotalSlash))
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

func (s *KeeperTestSuite) TestAfterRedelegationSlashed_Integration() {

	currHeight := int64(100)
	infractionHeight := currHeight
	currTime := time.Now()
	hctx := s.Ctx().WithBlockHeight(currHeight).WithBlockTime(currTime)
	slashFraction := sdkmath.LegacyOneDec()
	validatorTokens := sdkmath.NewIntWithDecimal(1, 9)
	undelegationBalance := sdkmath.OneInt() // just 1 unit, so we can focus more on the redelegation

	// Sanity check Governance account balance pre-slash
	govAccount := s.App.GovKeeper.GetGovernanceAccount(hctx).GetAddress()
	govAccountBalance := s.App.BankKeeper.GetBalance(hctx, govAccount, sdk.DefaultBondDenom)
	s.Require().True(govAccountBalance.IsZero())

	validators, err := s.App.StakingKeeper.GetAllValidators(s.Ctx())
	s.Require().NoError(err)

	// Sanity checks
	s.Require().True(validators[0].DelegatorShares.Equal(sdkmath.LegacyOneDec())) // 1 share at validator 0
	s.Require().True(validators[1].DelegatorShares.Equal(sdkmath.LegacyOneDec())) // 1 share at validator 1
	s.Require().True(validators[0].Tokens.Equal(validatorTokens))                 // 1e9 tokens at validator 0
	s.Require().True(validators[1].Tokens.Equal(validatorTokens))                 // 1e9 tokens at validator 1

	// Get validator1's delegator
	validator1Address, err := sdk.ValAddressFromBech32(validators[1].OperatorAddress)
	s.Require().NoError(err)
	validator1Delegation, err := s.App.StakingKeeper.GetValidatorDelegations(s.Ctx(), validator1Address)
	s.Require().NoError(err)
	validator1Delegator := validator1Delegation[0].DelegatorAddress

	// Redelegation from validator0 to validator1
	redelegation := stakingtypes.Redelegation{
		DelegatorAddress:    validator1Delegator,
		ValidatorSrcAddress: validators[0].OperatorAddress,
		ValidatorDstAddress: validators[1].OperatorAddress,
		Entries: []stakingtypes.RedelegationEntry{
			{
				CreationHeight:          currHeight + 1,                   // after current height
				CompletionTime:          currTime.Add(time.Hour),          // after current time
				InitialBalance:          sdkmath.NewIntWithDecimal(1, 20), // Large value to make sure we slash both the undelegation and redelegation
				SharesDst:               sdkmath.LegacyOneDec(),           // validator has 1 share
				UnbondingId:             0,                                // n/a
				UnbondingOnHoldRefCount: 0,                                // n/a
			},
		},
	}

	// Sanity check the NotBondedPoolName balance
	acc := s.App.AccountKeeper.GetModuleAccount(hctx, stakingtypes.NotBondedPoolName)
	s.Require().Empty(s.App.BankKeeper.GetAllBalances(hctx, acc.GetAddress()))

	// Mint the amount to be burned, due to the unbonding delegation, to the NotBondedPoolName, to avoid errors.
	expectedBurn := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, undelegationBalance))
	err = s.App.BankKeeper.MintCoins(hctx, minttypes.ModuleName, expectedBurn)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(
		hctx, minttypes.ModuleName, stakingtypes.NotBondedPoolName, expectedBurn,
	)
	s.Require().NoError(err)
	s.Require().Equal(expectedBurn, s.App.BankKeeper.GetAllBalances(hctx, acc.GetAddress()))

	// Unbonding delegation from validator1 - will get slashed first by SlashRedelegation
	undelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: validator1Delegator,
		ValidatorAddress: validators[1].OperatorAddress,
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CreationHeight:          currHeight + 1,          // after current height
				CompletionTime:          currTime.Add(time.Hour), // after current time
				InitialBalance:          sdkmath.Int{},           // n/a in this context
				Balance:                 undelegationBalance,     // Amount that will get slashed
				UnbondingId:             0,                       // n/a
				UnbondingOnHoldRefCount: 0,                       // n/a
			},
		},
	}
	err = s.App.StakingKeeper.SetUnbondingDelegation(hctx, undelegation)
	s.Require().NoError(err)

	_, err = s.App.StakingKeeper.SlashRedelegation(hctx, validators[0], redelegation, infractionHeight, slashFraction)
	s.Require().NoError(err)

	slashReport, found := s.App.ReportsKeeper.GetSlashReport(hctx, uint64(currHeight))
	s.Require().True(found)

	expectedTotalSlash := validators[1].Tokens.Add(undelegationBalance) // all validator and undelegation tokens slashed
	expectedSlashReport := types.SlashReport{
		Height: uint64(currHeight),
		Entries: []types.SlashEntry{
			{
				ValidatorAddress:          validators[1].OperatorAddress,
				DelegatorAddress:          validator1Delegator,
				DelegatorSlashAmount:      expectedTotalSlash,
				DelegatorBondedBalance:    sdkmath.ZeroInt(), // n/a
				DelegatorUnbondingBalance: sdkmath.ZeroInt(), // n/a
			},
		},
	}
	s.Require().Equal(expectedSlashReport, slashReport)

	// Sanity check unbonding delegation
	undelegationAfter, err := s.App.StakingKeeper.GetUnbondingDelegation(
		hctx, sdk.MustAccAddressFromBech32(validator1Delegator), validator1Address,
	)
	s.Require().NoError(err)
	s.Require().True(undelegationAfter.Entries[0].Balance.IsZero()) // the undelegation was fully slashed

	// Sanity check that tokens don't actually get burned
	govAccountBalance = s.App.BankKeeper.GetBalance(hctx, govAccount, sdk.DefaultBondDenom)
	s.Require().True(govAccountBalance.Amount.Equal(expectedTotalSlash))
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
