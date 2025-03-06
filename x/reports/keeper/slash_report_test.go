package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	for _, expectedSlashReport := range slashReports {
		actualSlashReport, found := testKeeper.GetSlashReport(ctx, expectedSlashReport.Height)
		require.True(t, found)
		require.Equal(t, keepertest.OrderSlashReportLexicographically(expectedSlashReport), actualSlashReport)
	}
}

func TestRemoveSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	for _, slashReport := range slashReports {
		// Report must be is state before deleted
		_, found := testKeeper.GetSlashReport(ctx, slashReport.Height)
		require.True(t, found)

		testKeeper.RemoveSlashReport(ctx, slashReport.Height)
		_, found = testKeeper.GetSlashReport(ctx, slashReport.Height)
		require.False(t, found)
	}
}

func TestGetAllSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	require.Equal(t,
		keepertest.OrderSlashReportsLexicographically(slashReports),
		testKeeper.GetAllSlashReport(ctx),
	)
}

func TestHasSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReport := keepertest.CreateNSlashReport(testKeeper, ctx, 1)[0]

	has := testKeeper.HasSlashReport(ctx, slashReport.Height)
	require.True(t, has)

	has = testKeeper.HasSlashReport(ctx, 2) // SlashReport in state has height 1
	require.False(t, has)
}

func TestUpdateSlashReportBalancesAtCurrentHeight(t *testing.T) {

	height := int64(100)
	heightU64 := uint64(height)
	otherHeight := int64(50)
	otherHeightU64 := uint64(otherHeight)

	entry0Val := sample.ValAddressBz()
	entry0Acc := sample.AccAddressBz()
	entry1Val := sample.ValAddressBz()
	entry1Acc := sample.AccAddressBz()

	entry0 := types.SlashEntry{
		ValidatorAddress:          entry0Val.String(),
		DelegatorAddress:          entry0Acc.String(),
		DelegatorSlashAmount:      math.NewInt(100),
		DelegatorBondedBalance:    math.ZeroInt(),
		DelegatorUnbondingBalance: math.ZeroInt(),
	}
	entry1 := types.SlashEntry{
		ValidatorAddress:          entry1Val.String(),
		DelegatorAddress:          entry1Acc.String(),
		DelegatorSlashAmount:      math.NewInt(200),
		DelegatorBondedBalance:    math.ZeroInt(),
		DelegatorUnbondingBalance: math.ZeroInt(),
	}

	setup := func() (keeper.Keeper, *testutil.MockStakingKeeper, sdk.Context) {
		// Set staking keeper mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		sk := testutil.NewMockStakingKeeper(ctrl)

		k, ctx := keepertest.ReportsKeeperWithKeepers(t, sk)
		ctx = ctx.WithBlockHeight(height)

		return k, sk, ctx
	}

	t.Run("update without any slash report", func(t *testing.T) {
		keeper, _, ctx := setup()

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// Slash report does not exist
		_, found := keeper.GetSlashReport(ctx, heightU64)
		require.False(t, found)
	})

	t.Run("update with slash report at other height", func(t *testing.T) {
		keeper, _, ctx := setup()

		// Set slash report at other height
		keeper.SetSlashReport(ctx, types.SlashReport{
			Height:  otherHeightU64,
			Entries: []types.SlashEntry{entry0, entry1},
		})

		// Slash report exists at the other height
		_, found := keeper.GetSlashReport(ctx, otherHeightU64)
		require.True(t, found)

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// Slash report does not exist at the height of interest
		_, found = keeper.GetSlashReport(ctx, heightU64)
		require.False(t, found)
	})

	t.Run("delegation returns unexpected error", func(t *testing.T) {
		keeper, sk, ctx := setup()

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.Delegation{}, fmt.Errorf("unexpected error"),
		)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoUnbondingDelegation,
		)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0},
		}
		keeper.SetSlashReport(ctx, report)
		// report = keepertest.OrderSlashReportLexicographically(report) // uncomment if you're going to use the report

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.ErrorContains(t, err, "unexpected error")
	})

	t.Run("undelegation returns unexpected error", func(t *testing.T) {
		keeper, sk, ctx := setup()

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.Delegation{}, stakingtypes.ErrNoDelegation,
		)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.UnbondingDelegation{}, fmt.Errorf("unexpected error"),
		)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0},
		}
		keeper.SetSlashReport(ctx, report)
		// report = keepertest.OrderSlashReportLexicographically(report) // uncomment if you're going to use the report

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.ErrorContains(t, err, "unexpected error")
	})

	t.Run("no delegation and no undelegation", func(t *testing.T) {
		keeper, sk, ctx := setup()

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.Delegation{}, stakingtypes.ErrNoDelegation,
		)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoUnbondingDelegation,
		)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0},
		}
		keeper.SetSlashReport(ctx, report)
		report = keepertest.OrderSlashReportLexicographically(report)

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// No changes
		reportAfter, found := keeper.GetSlashReport(ctx, uint64(height))
		require.True(t, found)
		require.EqualValues(t, report, reportAfter)
	})

	t.Run("delegation and no undelegation", func(t *testing.T) {
		keeper, sk, ctx := setup()

		entry0DelegatorShares := math.LegacyNewDec(123)
		entry0Delegation := stakingtypes.Delegation{
			DelegatorAddress: entry0Acc.String(),
			ValidatorAddress: entry0Val.String(),
			Shares:           entry0DelegatorShares,
		}

		validator := stakingtypes.Validator{
			Tokens:          math.NewInt(100),
			DelegatorShares: math.LegacyNewDec(200),
		}

		// We expect delegator to have (123/200)x100 tokens = 61.5
		tokens := validator.TokensFromSharesTruncated(entry0DelegatorShares)
		require.True(t, tokens.Equal(math.LegacyMustNewDecFromStr("61.5")))

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(entry0Delegation, nil)
		sk.EXPECT().GetValidator(ctx, entry0Val).Return(validator, nil)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoUnbondingDelegation,
		)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0},
		}
		keeper.SetSlashReport(ctx, report)
		report = keepertest.OrderSlashReportLexicographically(report)

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// Expected report updates
		report.Entries[0].DelegatorBondedBalance = math.NewInt(61) // 61.5 truncated to 61

		// Confirm changes
		reportAfter, found := keeper.GetSlashReport(ctx, uint64(height))
		require.True(t, found)
		require.EqualValues(t, report, reportAfter)
	})

	t.Run("undelegation and no delegation", func(t *testing.T) {
		keeper, sk, ctx := setup()

		// Total 6000 unbonding
		entry0UnbondingDelegation := stakingtypes.UnbondingDelegation{
			DelegatorAddress: entry0Acc.String(),
			ValidatorAddress: entry0Val.String(),
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{Balance: math.NewInt(1000)},
				{Balance: math.NewInt(2000)},
				{Balance: math.NewInt(3000)},
			},
		}

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(
			stakingtypes.Delegation{}, stakingtypes.ErrNoDelegation)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(entry0UnbondingDelegation, nil)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0},
		}
		keeper.SetSlashReport(ctx, report)
		report = keepertest.OrderSlashReportLexicographically(report)

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// Expected report updates
		report.Entries[0].DelegatorUnbondingBalance = math.NewInt(6000) // 1000+2000+3000

		// Confirm changes
		reportAfter, found := keeper.GetSlashReport(ctx, uint64(height))
		require.True(t, found)
		require.EqualValues(t, report, reportAfter)
	})

	t.Run("mix of delegations and undelegations", func(t *testing.T) {
		keeper, sk, ctx := setup()

		// Total 6000 unbonding for entry0
		entry0UnbondingDelegation := stakingtypes.UnbondingDelegation{
			DelegatorAddress: entry0Acc.String(),
			ValidatorAddress: entry0Val.String(),
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{Balance: math.NewInt(1000)},
				{Balance: math.NewInt(2000)},
				{Balance: math.NewInt(3000)},
			},
		}

		// Total 5000 unbonding for entry1
		entry1UnbondingDelegation := stakingtypes.UnbondingDelegation{
			DelegatorAddress: entry0Acc.String(),
			ValidatorAddress: entry0Val.String(),
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{Balance: math.NewInt(5000)},
			},
		}

		// Delegation corresponding to 61 tokens
		entry0DelegatorShares := math.LegacyNewDec(123)
		entry0Delegation := stakingtypes.Delegation{
			DelegatorAddress: entry0Acc.String(),
			ValidatorAddress: entry0Val.String(),
			Shares:           entry0DelegatorShares,
		}
		validator0 := stakingtypes.Validator{
			Tokens:          math.NewInt(100),
			DelegatorShares: math.LegacyNewDec(200),
		}

		// We expect delegator to have (123/200)x100 tokens = 61.5
		tokens0 := validator0.TokensFromSharesTruncated(entry0DelegatorShares)
		require.True(t, tokens0.Equal(math.LegacyMustNewDecFromStr("61.5")))

		// Delegation corresponding to 172 tokens
		entry1DelegatorShares := math.LegacyNewDec(345)
		entry1Delegation := stakingtypes.Delegation{
			DelegatorAddress: entry1Acc.String(),
			ValidatorAddress: entry1Val.String(),
			Shares:           entry1DelegatorShares,
		}
		validator1 := stakingtypes.Validator{
			Tokens:          math.NewInt(100),
			DelegatorShares: math.LegacyNewDec(200),
		}

		// We expect delegator to have (345/200)x100 tokens = 172.5
		tokens1 := validator1.TokensFromSharesTruncated(entry1DelegatorShares)
		require.True(t, tokens1.Equal(math.LegacyMustNewDecFromStr("172.5")))

		sk.EXPECT().GetDelegation(ctx, entry0Acc, entry0Val).Return(entry0Delegation, nil)
		sk.EXPECT().GetDelegation(ctx, entry1Acc, entry1Val).Return(entry1Delegation, nil)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry0Acc, entry0Val).Return(entry0UnbondingDelegation, nil)
		sk.EXPECT().GetUnbondingDelegation(ctx, entry1Acc, entry1Val).Return(entry1UnbondingDelegation, nil)
		sk.EXPECT().GetValidator(ctx, entry0Val).Return(validator0, nil)
		sk.EXPECT().GetValidator(ctx, entry1Val).Return(validator1, nil)

		// Set slash report at correct height
		report := types.SlashReport{
			Height:  heightU64,
			Entries: []types.SlashEntry{entry0, entry1},
		}
		keeper.SetSlashReport(ctx, report)
		report = keepertest.OrderSlashReportLexicographically(report)

		err := keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
		require.NoError(t, err)

		// Expected report updates
		if report.Entries[0].DelegatorAddress == entry0.DelegatorAddress { // report.Entries[0] is entry0
			report.Entries[0].DelegatorUnbondingBalance = math.NewInt(6000) // 1000+2000+3000
			report.Entries[1].DelegatorUnbondingBalance = math.NewInt(5000)
			report.Entries[0].DelegatorBondedBalance = math.NewInt(61)
			report.Entries[1].DelegatorBondedBalance = math.NewInt(172)
		} else if report.Entries[0].DelegatorAddress == entry1.DelegatorAddress { // report.Entries[0] is entry1
			report.Entries[0].DelegatorUnbondingBalance = math.NewInt(5000)
			report.Entries[1].DelegatorUnbondingBalance = math.NewInt(6000) // 1000+2000+3000
			report.Entries[0].DelegatorBondedBalance = math.NewInt(172)
			report.Entries[1].DelegatorBondedBalance = math.NewInt(61)
		} else {
			require.Fail(t, "unexpected case")
		}

		// Confirm changes
		reportAfter, found := keeper.GetSlashReport(ctx, uint64(height))
		require.True(t, found)
		require.EqualValues(t, report, reportAfter)
	})
}

func TestRemoveAllSlashReportsUntilHeight(t *testing.T) {

	startHeight := uint64(100)

	// Set heights to start from startHeight
	slashReports := keepertest.CreateNSlashReportWithoutStoring(10)
	for i := range slashReports {
		slashReports[i].Height = startHeight + uint64(i)
	}

	testCases := []struct {
		name               string
		setSlashReports    []types.SlashReport
		removeUntil        uint64
		expectSlashReports []types.SlashReport
	}{
		{
			name:               "Remove until < startHeight",
			setSlashReports:    slashReports,
			removeUntil:        startHeight - 1,
			expectSlashReports: slashReports,
		},
		{
			name:               "Remove until == startHeight",
			setSlashReports:    slashReports,
			removeUntil:        startHeight,
			expectSlashReports: slashReports[1:], // 1 removed
		},
		{
			name:               "Remove until is mid way through",
			setSlashReports:    slashReports,
			removeUntil:        startHeight + 4,
			expectSlashReports: slashReports[5:], // 5 removed
		},
		{
			name:               "One remaining",
			setSlashReports:    slashReports,
			removeUntil:        startHeight + 8,
			expectSlashReports: slashReports[9:], // 9 removed
		},
		{
			name:               "Remove until last height",
			setSlashReports:    slashReports,
			removeUntil:        startHeight + 9,
			expectSlashReports: nil, // All removed
		},
		{
			name:               "No reports == no pruning",
			setSlashReports:    nil,
			removeUntil:        startHeight,
			expectSlashReports: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testKeeper, ctx := keepertest.ReportsKeeper(t)

			// Set all slash reports
			for _, report := range tc.setSlashReports {
				testKeeper.SetSlashReport(ctx, report)
			}

			testKeeper.RemoveAllSlashReportsUntilHeight(ctx, tc.removeUntil)

			// Check remaining slash reports
			remainingReports := testKeeper.GetAllSlashReport(ctx)
			expectSlashReports := keepertest.OrderSlashReportsLexicographically(tc.expectSlashReports)
			require.EqualValues(t, expectSlashReports, remainingReports)
		})
	}
}

func TestPruneSlashReports(t *testing.T) {

	startHeight := uint64(100)

	// Set heights to start from startHeight
	n := 10
	slashReports := keepertest.CreateNSlashReportWithoutStoring(n)
	for i := range slashReports {
		slashReports[i].Height = startHeight + uint64(i)
	}

	testCases := []struct {
		name                    string
		setSlashReports         []types.SlashReport
		currentHeight           int64
		maxSlashReportAgeBlocks uint64
		expectSlashReports      []types.SlashReport
	}{
		{
			name:                    "Max age is large enough to keep all reports",
			setSlashReports:         slashReports,
			currentHeight:           110, // last report generated at the previous height
			maxSlashReportAgeBlocks: 11,
			expectSlashReports:      slashReports,
			// First report is at height 100. Remove until is 110-11 = 99. No reports removed.
		},
		{
			name:                    "First report is too old",
			setSlashReports:         slashReports,
			currentHeight:           110, // last report generated at the previous height
			maxSlashReportAgeBlocks: 10,
			expectSlashReports:      slashReports[1:],
			// First report is at height 100. Remove until is 110-10 = 100. 1 report removed.
		},
		{
			name: "Half of the reports are too old",

			setSlashReports:         slashReports,
			currentHeight:           110, // last report generated at the previous height
			maxSlashReportAgeBlocks: 6,
			expectSlashReports:      slashReports[5:],
			// First report is at height 100. Remove until is 110-6 = 104. 5 reports removed.
		},
		{
			name:                    "Only the last report is not too old",
			setSlashReports:         slashReports,
			currentHeight:           110, // last report generated at the previous height
			maxSlashReportAgeBlocks: 2,
			expectSlashReports:      slashReports[9:],
			// First report is at height 100. Remove until is 110-2 = 108. 9 reports removed.
		},
		{
			name:                    "All reports are too old",
			setSlashReports:         slashReports,
			currentHeight:           110, // last report generated at the previous height
			maxSlashReportAgeBlocks: 1,
			expectSlashReports:      nil,
			// First report is at height 100. Remove until is 110-1 = 109. All reports removed.
		},
		{
			name:                    "Last report is at current height, with max age 1 => last report retained",
			setSlashReports:         slashReports,
			currentHeight:           109, // last report generated at the current height
			maxSlashReportAgeBlocks: 1,
			expectSlashReports:      slashReports[9:],
			// First report is at height 100. Remove until is 109-1 = 108. 9 reports removed.
		},
		{
			name:                    "No reports => no pruning",
			setSlashReports:         nil,
			currentHeight:           110,
			maxSlashReportAgeBlocks: 10,
			expectSlashReports:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testKeeper, ctx := keepertest.ReportsKeeper(t)
			ctx = ctx.WithBlockHeight(tc.currentHeight)

			// Set param
			params := testKeeper.GetParams(ctx)
			params.MaxSlashReportAgeBlocks = tc.maxSlashReportAgeBlocks
			require.NoError(t, testKeeper.SetParams(ctx, params))

			// Set all slash reports
			for _, report := range tc.setSlashReports {
				testKeeper.SetSlashReport(ctx, report)
			}

			err := testKeeper.PruneSlashReports(ctx)
			require.NoError(t, err)

			// Check remaining slash reports
			remainingReports := testKeeper.GetAllSlashReport(ctx)
			expectSlashReports := keepertest.OrderSlashReportsLexicographically(tc.expectSlashReports)
			require.EqualValues(t, expectSlashReports, remainingReports)
		})
	}
}
