package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func TestSlashEntry_ValidateBasic(t *testing.T) {
	tests := []struct {
		name       string
		slashEntry types.SlashEntry
		expErrMsg  string
	}{
		{
			name:       "valid SlashEntry",
			slashEntry: testtypes.ValidSlashEntry1,
		},
		{
			name:       "valid SlashEntry - with zero DelegatorBondedBalance and DelegatorUnbondingBalance",
			slashEntry: testtypes.ValidSlashEntry3,
		},
		{
			name:       "invalid SlashEntry - validatorAddress cannot be non-valoper",
			slashEntry: testtypes.InvalidSlashEntryValidatorAddressNotValoper,
			expErrMsg:  "ValidatorAddress is not a valid Bech32 operator address",
		},
		{
			name:       "invalid SlashEntry - delegatorAddress cannot be a non-account address",
			slashEntry: testtypes.InvalidSlashEntryValidatorAddressNotAccAddress,
			expErrMsg:  "DelegatorAddress is not a valid Bech32 account address",
		},
		{
			name:       "invalid SlashEntry - DelegatorSlashAmount cannot be negative",
			slashEntry: testtypes.InvalidSlashEntryNegativeDelegatorSlashAmount,
			expErrMsg:  "DelegatorSlashAmount is not positive",
		},
		{
			name:       "invalid SlashEntry - DelegatorSlashAmount cannot be zero",
			slashEntry: testtypes.InvalidSlashEntryZeroDelegatorSlashAmount,
			expErrMsg:  "DelegatorSlashAmount is not positive",
		},
		{
			name:       "invalid SlashEntry - DelegatorBondedBalance cannot be negative",
			slashEntry: testtypes.InvalidSlashEntryNegativeDelegatorBondedBalance,
			expErrMsg:  "DelegatorBondedBalance cannot be negative",
		},
		{
			name:       "invalid SlashReport - DelegatorUnbondingBalance cannot be negative",
			slashEntry: testtypes.InvalidSlashEntryNegativeDelegatorUnbondingBalance,
			expErrMsg:  "DelegatorUnbondingBalance cannot be negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.slashEntry.ValidateBasic()
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
