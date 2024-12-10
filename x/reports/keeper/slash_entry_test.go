package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestGetSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntries := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 10)
	for _, expectedSlashEntry := range slashEntries {
		actualSlashEntry, found := testKeeper.GetSlashEntry(
			ctx, height, expectedSlashEntry.DelegatorAddress, expectedSlashEntry.ValidatorAddress,
		)
		require.True(t, found)
		require.Equal(t, expectedSlashEntry, actualSlashEntry)
	}
}

func TestRemoveSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntries := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 10)
	for _, slashEntry := range slashEntries {
		// Entry must be is state before deleted
		_, found := testKeeper.GetSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		require.True(t, found)

		testKeeper.RemoveSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		_, found = testKeeper.GetSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		require.False(t, found)
	}
}

func TestHasSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntry := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 1)[0]

	has := testKeeper.HasSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
	require.True(t, has)

	// SlashEntry in state has a different height
	has = testKeeper.HasSlashEntry(
		ctx, height+1, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress,
	)
	require.False(t, has)

	// SlashEntry in state has a different delegator address
	has = testKeeper.HasSlashEntry(
		ctx, height, "bad_delegator_address", slashEntry.ValidatorAddress,
	)
	require.False(t, has)

	// SlashEntry in state has a different validator address
	has = testKeeper.HasSlashEntry(
		ctx, height, slashEntry.DelegatorAddress, "bad_validator_address",
	)
	require.False(t, has)
}

func (s *KeeperTestSuite) TestInsertSlashEntry() {
	height := int64(10)
	delegatorAddress := testtypes.TestSeqAddr1
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
	}{
		{
			name:        "slash entry set correctly first time",
			height:      height,
			delAddr:     delegatorAddress,
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
				DelegatorBondedBalance:    sdkmath.OneInt(), // Set to one to make sure that the value updates to 0
				DelegatorUnbondingBalance: sdkmath.OneInt(), // Set to one to make sure that the value updates to 0
			},
			height:      height,
			delAddr:     delegatorAddress,
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
			name:        "errors if context height is zero",
			height:      0,
			delAddr:     delegatorAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg:   "slashing height must be positive, received: 0",
		},
		{
			name:        "errors if context height is negative",
			height:      -1,
			delAddr:     delegatorAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.OneInt(),
			expErrMsg:   "slashing height must be positive, received: -1",
		},
		{
			name:        "errors if slashed amount is zero",
			height:      height,
			delAddr:     delegatorAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.ZeroInt(),
			expErrMsg:   "slash amount must be positive, received 0",
		},
		{
			name:        "errors if slashed amount is negative",
			height:      height,
			delAddr:     delegatorAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.NewInt(-1),
			expErrMsg:   "slash amount must be positive, received -1",
		},
		// The test case below was only written for the sake of completion. We should never be in a situation where a
		// slash entry is constructed in an invalid manner. In fact to trigger this case we will be storing a negative
		// slash amount in state, which should never occur if we call validate basic before storing.
		{
			name: "errors if constructed invalid slash entry",
			setSlashEntry: &types.SlashEntry{
				ValidatorAddress: testtypes.TestValAddr1Str,
				DelegatorAddress: testtypes.TestSeqAddr1Str,

				// Set to a |value| > slash amount so that the constructed slash entry has a negative slash amount
				DelegatorSlashAmount: sdkmath.NewInt(-10),

				DelegatorBondedBalance:    sdkmath.ZeroInt(),
				DelegatorUnbondingBalance: sdkmath.ZeroInt(),
			},
			height:      height,
			delAddr:     delegatorAddress,
			valAddr:     validatorAddress,
			slashAmount: sdkmath.NewInt(1),
			expErrMsg:   "constructed invalid slash entry",
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

			// Execute InsertSlashEntry
			err := s.App.ReportsKeeper.InsertSlashEntry(ctxWithHeight, tc.valAddr, tc.delAddr, tc.slashAmount)
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
		})
	}
}
